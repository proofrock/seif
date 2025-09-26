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
	"os"
	"seif/params"
	"strconv"
)

func Parse() {
	_db := getEnvOrDefault("SEIF_DB", "./seif.db")
	_port := getEnvIntOrDefault("SEIF_PORT", 34543)
	_maxDays := getEnvIntOrDefault("SEIF_MAX_DAYS", 3)
	_defaultDays := getEnvIntOrDefault("SEIF_DEFAULT_DAYS", 3)
	_maxBytes := getEnvIntOrDefault("SEIF_MAX_BYTES", 1024)

	flag.Parse()

	params.DbPath = _db
	params.Port = _port
	params.MaxDays = _maxDays
	params.DefaultDays = min(_defaultDays, _maxDays)
	params.MaxBytes = _maxBytes
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		} else {
			println("FATAL: Env var '", key, "' is not numeric")
		}
	}
	return defaultValue
}
