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
	"fmt"
	"seif/params"
	"seif/utils"
)

const DB_VERSION = 1

const SQL_CREATE = `
 	CREATE TABLE SECRETS (
		ID TEXT PRIMARY KEY,
		IV TEXT,
		SECRET TEXT,
		SHA TEXT,
		EXPIRY INTEGER,
		TS TEXT
	)`

var SQL_CREATE_2 = fmt.Sprintf("CREATE VIEW VERSION AS SELECT %d AS VERSION", DB_VERSION)

func InitDb() {
	// Execute non-concurrently
	params.Lock.Lock()
	defer params.Lock.Unlock()

	if _, err := params.Db.Exec(SQL_CREATE); err != nil {
		utils.Abort("in creating db: %s", err)
	}

	if _, err := params.Db.Exec(SQL_CREATE_2); err != nil {
		utils.Abort("in creating db: %s", err)
	}
}
