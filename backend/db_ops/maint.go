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
	"os"
	"seif/params"
	"seif/utils"
	"time"
)

const maint_period = 5 // min

const SQL_MAINT = "DELETE FROM SECRETS WHERE TS < DATETIME('now', '-' || EXPIRY || ' days')"

func maint(allowToPanic bool) {
	// Execute non-concurrently
	params.Lock.Lock()
	defer params.Lock.Unlock()

	if _, err := params.Db.Exec(SQL_MAINT); err != nil {
		if allowToPanic {
			utils.Abort("in doing cleanup: %s\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "in doing maintenance cleanup: %s\n", err.Error())
		}
	}

	if _, err := params.Db.Exec("VACUUM"); err != nil {
		if allowToPanic {
			utils.Abort("in doing vacuum: %s\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "in doing maintenance vacuum: %s\n", err.Error())
		}
	}
}

func StartMaint() {
	maint(true)
	for range time.Tick((maint_period * time.Minute)) {
		maint(false)
	}
}
