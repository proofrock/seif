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
package db_ops

import (
	"encoding/json"
	"fmt"
	"os"
	"seif/params"
	"seif/utils"
	"time"

	"go.etcd.io/bbolt"
)

const maint_period = 5 // min

type SecretData struct {
	Secret []byte `json:"secret"`
	Expiry int    `json:"expiry"`
	TS     int64  `json:"ts"`
}

func maint(allowToPanic bool) {
	// Execute non-concurrently
	params.Lock.Lock()
	defer params.Lock.Unlock()

	now := time.Now().Unix()
	var keysToDelete [][]byte

	err := params.Db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(SECRETS_BUCKET)
		if bucket == nil {
			return nil
		}

		// Collect expired keys
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var secretData SecretData
			if err := json.Unmarshal(v, &secretData); err != nil {
				continue // Skip malformed data
			}

			// Check if expired (ts + expiry*days in seconds)
			expirationTime := secretData.TS + int64(secretData.Expiry*24*60*60)
			if now > expirationTime {
				// Copy key since cursor data is only valid during iteration
				keyCopy := make([]byte, len(k))
				copy(keyCopy, k)
				keysToDelete = append(keysToDelete, keyCopy)
			}
		}

		// Delete expired keys
		for _, key := range keysToDelete {
			if err := bucket.Delete(key); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if allowToPanic {
			utils.Abort("in doing cleanup: %s\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "in doing maintenance cleanup: %s\n", err.Error())
		}
	}

	if len(keysToDelete) > 0 {
		fmt.Printf("Cleaned up %d expired secrets\n", len(keysToDelete))
	}
}

func StartMaint() {
	maint(true)
	for range time.Tick((maint_period * time.Minute)) {
		maint(false)
	}
}
