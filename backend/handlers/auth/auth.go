/*
 * Copyright (C) 2024- Germano Rizzo
 *
 * This file is part of Seif.
 *
 * Seif is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Seif is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with Seif.  If not, see <http://www.gnu.org/licenses/>.
 */
package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"seif/middleware"
	"seif/oauth2"
	"seif/utils"
	"time"

	stdoauth2 "golang.org/x/oauth2"
)

type StateData struct {
	Expiry       time.Time
	CodeVerifier string
}

var stateStore = make(map[string]*StateData)

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !oauth2.OAuth2Config.Enabled {
		utils.SendError(w, http.StatusServiceUnavailable, utils.FHE001, "OAuth2 not enabled", nil)
		return
	}

	state, err := oauth2.GenerateStateToken()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE007, "state generation", &err)
		return
	}

	// Generate PKCE parameters
	codeVerifier, err := oauth2.GenerateCodeVerifier()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE007, "code verifier generation", &err)
		return
	}

	// Store state with expiration and code verifier
	stateStore[state] = &StateData{
		Expiry:       time.Now().Add(10 * time.Minute),
		CodeVerifier: codeVerifier,
	}

	// Clean up expired states
	cleanupExpiredStates()

	// Create authorization URL with PKCE
	url := oauth2.OAuth2Config.Config.AuthCodeURL(state, stdoauth2.S256ChallengeOption(codeVerifier))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func Callback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !oauth2.OAuth2Config.Enabled {
		utils.SendError(w, http.StatusServiceUnavailable, utils.FHE001, "OAuth2 not enabled", nil)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	oauthError := r.URL.Query().Get("error")
	oauthErrorDescription := r.URL.Query().Get("error_description")

	// Check for OAuth error response
	if oauthError != "" {
		log.Printf("OAuth2: Error from provider: %s - %s", oauthError, oauthErrorDescription)
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "OAuth provider error: "+oauthError, nil)
		return
	}

	// Validate required parameters
	if code == "" {
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "missing authorization code", nil)
		return
	}

	// Validate state
	stateData, exists := stateStore[state]
	if !exists || time.Now().After(stateData.Expiry) {
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "invalid state", nil)
		return
	}

	// Get code verifier for PKCE
	codeVerifier := stateData.CodeVerifier

	// Clean up used state
	delete(stateStore, state)

	// Exchange code for token with PKCE
	token, err := oauth2.OAuth2Config.Config.Exchange(r.Context(), code, stdoauth2.VerifierOption(codeVerifier))
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "token exchange", &err)
		return
	}

	// Get user info
	userInfo, err := oauth2.GetUserInfo(r.Context(), token)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE007, "user info", &err)
		return
	}

	log.Printf("OAuth2: User authenticated: %s (%s)", userInfo.Name, userInfo.Email)

	// Check email whitelist
	if !oauth2.IsEmailAllowed(userInfo.Email) {
		log.Printf("OAuth2: Access denied for email: %s (not in whitelist)", userInfo.Email)

		// Redirect to error page with user-friendly message
		errorHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Access Denied - Seif</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 20px;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 10px;
            padding: 40px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            text-align: center;
            max-width: 500px;
        }
        .error-icon { font-size: 48px; margin-bottom: 20px; }
        h1 { color: #e74c3c; margin-bottom: 20px; }
        p { color: #666; line-height: 1.6; margin-bottom: 30px; }
        .email { font-family: monospace; background: #f8f9fa; padding: 5px 10px; border-radius: 5px; }
        .button {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 12px 24px;
            text-decoration: none;
            border-radius: 5px;
            transition: background 0.3s;
        }
        .button:hover { background: #5a67d8; }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">🚫</div>
        <h1>Access Denied</h1>
        <p>Your email address <span class="email">` + userInfo.Email + `</span> is not authorized to access this application.</p>
        <p>Please contact your administrator if you believe this is an error.</p>
        <a href="/" class="button">← Back to Home</a>
    </div>
</body>
</html>`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(errorHTML))
		return
	}

	// Create session
	sessionID := middleware.CreateSession(userInfo, token)
	log.Printf("OAuth2: Session created: %s", sessionID)

	// Set session cookie
	cookie := &http.Cookie{
		Name:     "seif_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   24 * 60 * 60, // 24 hours
	}
	http.SetCookie(w, cookie)

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	if cookie, err := r.Cookie("seif_session"); err == nil {
		middleware.DestroySession(cookie.Value)
	}

	// Clear session cookie
	cookie := &http.Cookie{
		Name:     "seif_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "logged out"})
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if !oauth2.OAuth2Config.Enabled {
		json.NewEncoder(w).Encode(map[string]interface{}{"user": nil})
		return
	}

	// Check for session cookie manually since this endpoint isn't protected by middleware
	cookie, err := r.Cookie("seif_session")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"user": nil})
		return
	}

	// Get session data from middleware package
	sessionData := middleware.GetSessionData(cookie.Value)
	if sessionData == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"user": nil})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"user": sessionData.UserInfo})
}

func cleanupExpiredStates() {
	now := time.Now()
	for state, stateData := range stateStore {
		if now.After(stateData.Expiry) {
			delete(stateStore, state)
		}
	}
}