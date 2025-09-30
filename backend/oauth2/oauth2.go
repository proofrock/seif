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
package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

type Config struct {
	Enabled        bool
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	EmailWhitelist []string
	AllowGuestLink bool
	Config         *oauth2.Config
}

type UserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

var OAuth2Config *Config

func init() {
	OAuth2Config = LoadConfig()
}

func LoadConfig() *Config {
	enabled := strings.ToLower(os.Getenv("SEIF_OAUTH_ENABLED")) == "true"
	allowGuestLink := strings.ToLower(os.Getenv("SEIF_ALLOW_GUEST_LINK")) == "true"

	// Parse email whitelist
	var emailWhitelist []string
	whitelistEnv := os.Getenv("SEIF_OAUTH_EMAIL_WHITELIST")
	if whitelistEnv != "" {
		emailWhitelist = strings.Split(whitelistEnv, ",")
		// Trim whitespace from each email
		for i, email := range emailWhitelist {
			emailWhitelist[i] = strings.TrimSpace(email)
		}
	}

	config := &Config{
		Enabled:        enabled,
		ClientID:       os.Getenv("SEIF_OAUTH_CLIENT_ID"),
		ClientSecret:   os.Getenv("SEIF_OAUTH_CLIENT_SECRET"),
		RedirectURL:    os.Getenv("SEIF_OAUTH_REDIRECT_URI"),
		EmailWhitelist: emailWhitelist,
		AllowGuestLink: allowGuestLink,
	}

	if !enabled {
		log.Println("OAuth2: Disabled (SEIF_OAUTH_ENABLED is not set to 'true')")
		return config
	}

	// Check basic OAuth2 configuration
	authURL := os.Getenv("SEIF_OAUTH_AUTH_URL")
	tokenURL := os.Getenv("SEIF_OAUTH_TOKEN_URL")
	userInfoURL := os.Getenv("SEIF_OAUTH_USERINFO_URL")

	if config.ClientID == "" || config.ClientSecret == "" || config.RedirectURL == "" || authURL == "" || tokenURL == "" || userInfoURL == "" {
		log.Println("OAuth2: Disabled due to missing configuration:")
		if config.ClientID == "" {
			log.Println("  - SEIF_OAUTH_CLIENT_ID is required")
		}
		if config.ClientSecret == "" {
			log.Println("  - SEIF_OAUTH_CLIENT_SECRET is required")
		}
		if config.RedirectURL == "" {
			log.Println("  - SEIF_OAUTH_REDIRECT_URI is required")
		}
		if authURL == "" {
			log.Println("  - SEIF_OAUTH_AUTH_URL is required")
		}
		if tokenURL == "" {
			log.Println("  - SEIF_OAUTH_TOKEN_URL is required")
		}
		if userInfoURL == "" {
			log.Println("  - SEIF_OAUTH_USERINFO_URL is required")
		}
		log.Println("  - SEIF_OAUTH_SCOPES is optional (defaults to 'openid email profile')")
		config.Enabled = false
		return config
	}

	// Set up OAuth2 endpoint with custom URLs
	endpoint := oauth2.Endpoint{
		AuthURL:  authURL,
		TokenURL: tokenURL,
	}

	// Configure scopes
	scopesEnv := os.Getenv("SEIF_OAUTH_SCOPES")
	var scopes []string
	if scopesEnv != "" {
		scopes = strings.Split(scopesEnv, " ")
	} else {
		scopes = []string{"openid", "email", "profile"}
	}

	config.Config = &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Scopes:       scopes,
		Endpoint:     endpoint,
	}

	guestLinkStatus := "disabled"
	if config.AllowGuestLink {
		guestLinkStatus = "enabled"
	}

	if len(config.EmailWhitelist) > 0 {
		log.Printf("OAuth2: Enabled with custom provider (email whitelist: %v, guest links: %s)", config.EmailWhitelist, guestLinkStatus)
	} else {
		log.Printf("OAuth2: Enabled with custom provider (all emails allowed, guest links: %s)", guestLinkStatus)
	}
	return config
}

func GenerateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func GenerateCodeVerifier() (string, error) {
	return oauth2.GenerateVerifier(), nil
}

func GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	if !OAuth2Config.Enabled {
		return nil, fmt.Errorf("OAuth2 is not enabled")
	}

	client := OAuth2Config.Config.Client(ctx, token)

	userInfoURL := os.Getenv("SEIF_OAUTH_USERINFO_URL")
	if userInfoURL == "" {
		return nil, fmt.Errorf("SEIF_OAUTH_USERINFO_URL is required")
	}

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func IsEmailAllowed(email string) bool {
	if !OAuth2Config.Enabled {
		return true
	}

	// If no whitelist is configured, allow all emails
	if len(OAuth2Config.EmailWhitelist) == 0 {
		return true
	}

	// Check if email is in whitelist
	email = strings.ToLower(strings.TrimSpace(email))
	for _, allowedEmail := range OAuth2Config.EmailWhitelist {
		if strings.ToLower(allowedEmail) == email {
			return true
		}
	}

	return false
}
