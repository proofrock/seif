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
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"seif/oauth2"
	"time"

	oauth2lib "golang.org/x/oauth2"
)

type contextKey string

const UserContextKey contextKey = "user"

type SessionStore struct {
	sessions map[string]*SessionData
}

type SessionData struct {
	UserInfo  *oauth2.UserInfo
	Token     *oauth2lib.Token
	ExpiresAt time.Time
}

var sessions = &SessionStore{
	sessions: make(map[string]*SessionData),
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If OAuth2 is not enabled, skip authentication
		if !oauth2.OAuth2Config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Check for session cookie
		cookie, err := r.Cookie("seif_session")
		if err != nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Validate session
		sessionData, exists := sessions.sessions[cookie.Value]
		if !exists || time.Now().After(sessionData.ExpiresAt) {
			// Clean up expired session
			if exists {
				delete(sessions.sessions, cookie.Value)
			}
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, sessionData.UserInfo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CreateSession(userInfo *oauth2.UserInfo, token *oauth2lib.Token) string {
	sessionID := generateSessionID()
	sessions.sessions[sessionID] = &SessionData{
		UserInfo:  userInfo,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
	}
	return sessionID
}

func DestroySession(sessionID string) {
	delete(sessions.sessions, sessionID)
}

func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func GetUserFromContext(ctx context.Context) *oauth2.UserInfo {
	if user, ok := ctx.Value(UserContextKey).(*oauth2.UserInfo); ok {
		return user
	}
	return nil
}

func GetSessionData(sessionID string) *SessionData {
	sessionData, exists := sessions.sessions[sessionID]
	if !exists || time.Now().After(sessionData.ExpiresAt) {
		// Clean up expired session
		if exists {
			delete(sessions.sessions, sessionID)
		}
		return nil
	}
	return sessionData
}
