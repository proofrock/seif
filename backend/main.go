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
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"seif/db_ops"
	"seif/flags"
	"seif/handlers/auth"
	"seif/handlers/get_init_data"
	"seif/handlers/get_secret"
	"seif/handlers/get_secret_status"
	"seif/handlers/ping"
	"seif/handlers/put_secret"
	"seif/middleware"
	"seif/params"
	"seif/utils"

	"go.etcd.io/bbolt"
)

//go:embed static/*
var static embed.FS

func main() {
	flags.Parse()

	dbIsNew := !utils.FileExists(params.DbPath)

	// FIXME don't open a new connection for each operation
	var err error
	params.Db, err = bbolt.Open(params.DbPath, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer params.Db.Close()

	// Populates db
	if dbIsNew {
		db_ops.InitDb()
	}

	// Maintenance
	go db_ops.StartMaint()

	// Setup routes
	mux := http.NewServeMux()

	// Static files
	staticFS, _ := fs.Sub(static, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	// API routes
	mux.HandleFunc("/api/ping", ping.Ping)
	mux.HandleFunc("/api/getInitData", get_init_data.GetInitData)
	mux.HandleFunc("/api/getSecret", get_secret.GetSecret)
	mux.HandleFunc("/api/getSecretStatus", get_secret_status.GetSecretStatus)
	mux.HandleFunc("/api/putSecret", middleware.RequireAuth(put_secret.PutSecret))

	// OAuth2 routes
	mux.HandleFunc("/api/auth/login", auth.Login)
	mux.HandleFunc("/api/auth/callback", auth.Callback)
	mux.HandleFunc("/api/auth/logout", auth.Logout)
	mux.HandleFunc("/api/auth/user", auth.GetUser)
	mux.HandleFunc("/api/auth/generate-guest-link", middleware.RequireAuth(auth.GenerateGuestLink))

	// Create server with custom header
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", params.Port),
		Handler: addServerHeader(mux),
	}

	log.Println("Started server on port", params.Port)
	log.Printf("Everything ok. Please open http://localhost:%d\n", params.Port)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func addServerHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "seif v."+params.VERSION)
		next.ServeHTTP(w, r)
	})
}
