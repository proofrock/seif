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
package flags

import (
	"flag"
	"seif/params"
)

func Parse() {
	_db := flag.String("db", "./seif.db", "The path of the sqlite database")
	_port := flag.Int("port", 34543, "Port")
	_maxDays := flag.Int("max-days", 3, "Maximum retention days to allow")
	_maxBytes := flag.Int("max-bytes", 1024, "Maximum size, in bytes, of a secret")

	flag.Parse()

	params.DbPath = *_db
	params.Port = *_port
	params.MaxDays = *_maxDays
	params.MaxBytes = *_maxBytes
}
