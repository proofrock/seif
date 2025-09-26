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
package put_secret

import (
	"encoding/json"
	"fmt"
	"net/http"
	"seif/crypton"
	"seif/params"
	"seif/utils"
)

type request struct {
	Secret string `json:"secret"`
	Expiry int    `json:"expiry"`
}

type response struct {
	Id  string `json:"id"`
	Key string `json:"key"`
}

const SQL = `
	INSERT INTO SECRETS (ID, SECRET, EXPIRY, TS)
	VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`

func PutSecret(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := new(request)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		utils.SendError(w, http.StatusBadRequest, utils.FHE004, "body", &err)
		return
	}

	if len(req.Secret) > params.MaxBytes {
		utils.SendError(w, http.StatusBadRequest, utils.FHE005, "", nil)
		return
	}

	if req.Expiry < 1 || req.Expiry > params.MaxDays {
		utils.SendError(w, http.StatusBadRequest, utils.FHE006, fmt.Sprint(params.MaxDays), nil)
		return
	}

	id, key, crypto, err := crypton.Encode(req.Secret)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE007, "", &err)
		return
	}

	ret := response{Id: crypton.Bs2str(id), Key: crypton.Bs2str(key)}

	params.Lock.Lock()
	defer params.Lock.Unlock()

	_, err = params.Db.Exec(SQL, ret.Id, crypto, req.Expiry)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, utils.FHE002, "secrets", &err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
