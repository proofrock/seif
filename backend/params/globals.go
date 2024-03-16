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
package params

import (
	"database/sql"
	"fmt"
	"sync"
)

const VERSION = "v0.0.0"

// https://manytools.org/hacker-tools/ascii-banner/, profile "Small Slant"
const banner = `   ____    _ ___
  / __/__ (_) _/
 _\ \/ -_) / _/ 
/___/\__/_/_/`

var Lock sync.Mutex

var Db *sql.DB

func init() {
	fmt.Println(banner, VERSION)
	fmt.Println()
}
