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
package get_secret_status

import (
	"encoding/json"
	"net/http"
	"seif/params"
	"seif/utils"
)

type response struct {
	Pristine bool `json:"pristine"`
}

const SQL_GET_SECRET = "SELECT 1 FROM SECRETS WHERE ID = $1"

func GetSecretStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")

	params.Lock.Lock()
	defer params.Lock.Unlock()

	rows, err := params.Db.Query(SQL_GET_SECRET, id)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE001, "secret", &err)
		return
	}
	defer rows.Close()

	ret := rows.Next()
	if err = rows.Err(); err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE003, "secret", &err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response{Pristine: ret})
}
