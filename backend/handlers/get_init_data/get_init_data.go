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
package get_init_data

import (
	"encoding/json"
	"net/http"
	"seif/oauth2"
	"seif/params"
)

type response struct {
	Version      string `json:"version"`
	MaxDays      int    `json:"max_days"`
	DefaultDays  int    `json:"default_days"`
	OAuthEnabled bool   `json:"oauth_enabled"`
}

func GetInitData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := response{
		Version:      params.VERSION,
		MaxDays:      params.MaxDays,
		DefaultDays:  params.DefaultDays,
		OAuthEnabled: oauth2.OAuth2Config.Enabled,
	}

	json.NewEncoder(w).Encode(resp)
}
