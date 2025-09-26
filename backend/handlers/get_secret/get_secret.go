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
package get_secret

import (
	"context"
	"encoding/json"
	"net/http"
	"seif/crypton"
	"seif/params"
	"seif/utils"
)

type response struct {
	Secret *string `json:"secret"`
}

const SQL1 = "SELECT SECRET FROM SECRETS WHERE ID = $1"
const SQL2 = "DELETE FROM SECRETS WHERE ID = $1"

func GetSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	idBs, err := crypton.Str2bs(id)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "id", &err)
		return
	}

	key := r.URL.Query().Get("key")
	keyBs, err := crypton.Str2bs(key)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "key", &err)
		return
	}

	params.Lock.Lock()
	defer params.Lock.Unlock()

	ret := response{}

	tx, err := params.Db.BeginTx(context.Background(), nil)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "FHE007", "", &err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(SQL1, id)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE001, "secret", &err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var secret []byte
		err = rows.Scan(&secret)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, utils.FHE001, "secret", &err)
			return
		}

		plaintxt, err := crypton.Decode(idBs, keyBs, secret)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, utils.FHE008, "decryption", &err)
			return
		}

		ret.Secret = &plaintxt
	}
	if err = rows.Err(); err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE003, "secret", &err)
		return
	}

	if ret.Secret != nil {
		_, err := tx.Exec(SQL2, id)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, utils.FHE009, "secret", &err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE009, "transaction", &err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
