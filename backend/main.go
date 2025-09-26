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
package main

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"seif/db_ops"
	"seif/flags"
	"seif/handlers/get_init_data"
	"seif/handlers/get_secret"
	"seif/handlers/get_secret_status"
	"seif/handlers/ping"
	"seif/handlers/put_secret"
	"seif/params"
	"seif/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "modernc.org/sqlite"
)

//go:embed static/*
var static embed.FS

func main() {
	flags.Parse()

	dbIsNew := !utils.FileExists(params.DbPath)

	// FIXME don't open a new connection for each operation
	var err error
	params.Db, err = sql.Open("sqlite", params.DbPath)
	if err != nil {
		panic(err)
	}
	defer params.Db.Close()

	// Populates db
	if dbIsNew {
		db_ops.InitDb()
	} else {
		// check db version

		row := params.Db.QueryRow("SELECT VERSION FROM VERSION")
		var dbVersion int
		if err := row.Scan(&dbVersion); err != nil {
			panic(err)
		}
		if dbVersion != db_ops.DB_VERSION {
			utils.Abort("DB version is %d but should be %d. Please upgrade the database or the application.", dbVersion, db_ops.DB_VERSION)
		}
	}

	// Maintenance

	go db_ops.StartMaint()

	// server

	app := fiber.New(fiber.Config{ServerHeader: "seif v." + params.VERSION, AppName: "seif", DisableStartupMessage: true})

	app.Use(recover.New())

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
	}))

	app.Get("/api/ping", ping.Ping)
	app.Get("/api/getInitData", get_init_data.GetInitData)
	app.Delete("/api/getSecret", get_secret.GetSecret)
	app.Get("/api/getSecretStatus", get_secret_status.GetSecretStatus)
	app.Put("/api/putSecret", put_secret.PutSecret)

	fmt.Println("  - server on port", params.Port)
	fmt.Printf("  - all ok. Please open http://localhost:%d\n", params.Port)
	app.Listen(fmt.Sprintf(":%d", params.Port))
}
