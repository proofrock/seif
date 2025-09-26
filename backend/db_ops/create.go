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
	"encoding/binary"
	"seif/params"
	"seif/utils"

	"go.etcd.io/bbolt"
)

const DB_VERSION = 1

var SECRETS_BUCKET = []byte("secrets")
var VERSION_BUCKET = []byte("version")

func InitDb() {
	// Execute non-concurrently
	params.Lock.Lock()
	defer params.Lock.Unlock()

	err := params.Db.Update(func(tx *bbolt.Tx) error {
		// Create secrets bucket
		_, err := tx.CreateBucket(SECRETS_BUCKET)
		if err != nil {
			return err
		}

		// Create version bucket and set version
		versionBucket, err := tx.CreateBucket(VERSION_BUCKET)
		if err != nil {
			return err
		}

		// Store version as binary
		versionBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(versionBytes, DB_VERSION)
		return versionBucket.Put([]byte("version"), versionBytes)
	})

	if err != nil {
		utils.Abort("in creating db: %s", err)
	}
}

func GetDBVersion() (int, error) {
	var version int
	err := params.Db.View(func(tx *bbolt.Tx) error {
		versionBucket := tx.Bucket(VERSION_BUCKET)
		if versionBucket == nil {
			return nil // DB not initialized
		}

		versionBytes := versionBucket.Get([]byte("version"))
		if versionBytes == nil {
			return nil
		}

		version = int(binary.LittleEndian.Uint32(versionBytes))
		return nil
	})

	return version, err
}
