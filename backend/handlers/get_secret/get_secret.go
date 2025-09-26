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
	"encoding/json"
	"net/http"
	"seif/crypton"
	"seif/db_ops"
	"seif/params"
	"seif/utils"

	"go.etcd.io/bbolt"
)

type response struct {
	Secret *string `json:"secret"`
}

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

	err = params.Db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(db_ops.SECRETS_BUCKET)
		if bucket == nil {
			return nil // Secret doesn't exist
		}

		// Get the secret data
		secretBytes := bucket.Get([]byte(id))
		if secretBytes == nil {
			return nil // Secret doesn't exist
		}

		// Unmarshal secret data
		var secretData db_ops.SecretData
		if err := json.Unmarshal(secretBytes, &secretData); err != nil {
			return err
		}

		// Decrypt the secret
		plaintxt, err := crypton.Decode(idBs, keyBs, secretData.Secret)
		if err != nil {
			return err
		}

		ret.Secret = &plaintxt

		// Delete the secret (one-time use)
		return bucket.Delete([]byte(id))
	})

	if err != nil {
		// Check if it's a decryption error
		if ret.Secret == nil {
			utils.SendError(w, http.StatusBadRequest, utils.FHE008, "decryption", &err)
		} else {
			utils.SendError(w, http.StatusInternalServerError, utils.FHE009, "secret", &err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ret)
}
