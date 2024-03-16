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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"seif/params"
	"sort"
	"strings"
	"time"
)

const bkpTimeFormat = "060102-150405"

var bkpTimeGlob = strings.Repeat("?", len(bkpTimeFormat))

const bkpFile = "seif_%s.db"
const numFiles = 8

func Backup() {
	var bkpDir = filepath.Join(filepath.Dir(params.DbPath), "backups")
	var err error

	if _, err = os.Stat(bkpDir); errors.Is(err, os.ErrNotExist) {
		if err = os.Mkdir(bkpDir, 0755); err != nil {
			panic(err)
		}
	}

	// Execute non-concurrently
	params.Lock.Lock()
	defer params.Lock.Unlock()

	now := time.Now().Format(bkpTimeFormat)
	fname := fmt.Sprintf(filepath.Join(bkpDir, bkpFile), now)
	_, err = params.Db.Exec("VACUUM INTO ?", fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "backup: %s\n", err.Error())
		return
	}

	// delete the backup files, except for the last n
	list, err := filepath.Glob(fmt.Sprintf(filepath.Join(bkpDir, bkpFile), bkpTimeGlob))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sched. task (pruning bkp files): %s\n", err.Error())
		return
	}

	sort.Strings(list)
	for i := 0; i < len(list)-numFiles; i++ {
		os.Remove(list[i])
	}
}
