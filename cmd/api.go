// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.
package main

import (
	"context"
	"fmt"
	"log"

	u "github.com/iliafrenkel/go-pb/src/api/auth/sqldb"
	"github.com/iliafrenkel/go-pb/src/api/http"
	p "github.com/iliafrenkel/go-pb/src/api/paste/sqldb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var apiServer *http.APIServer
var db *gorm.DB

// startAPIServer connects to the database, initialises User and Paste
// services and starts the API server  on the provided address.
func startAPIServer(opts http.APIServerOptions) error {
	// Connect to the database
	var err error
	db, err = gorm.Open(postgres.Open(opts.DBConnectionString), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to establish database connection: %w", err)
	}

	// Create UserService
	userSvc, err := u.New(u.SvcOptions{
		DBConnection:  db,
		DBAutoMigrate: opts.DBAutoMigrate,
		TokenSecret:   opts.TokenSecret,
	})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to create UserService: %w", err)
	}

	// Create PasteService
	pasteSvc, err := p.New(p.SvcOptions{
		DBConnection:  db,
		DBAutoMigrate: opts.DBAutoMigrate,
	})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to create PasteService: %w", err)
	}

	apiServer = http.New(pasteSvc, userSvc, opts)

	log.Println("API server listening on ", opts.Addr)

	return apiServer.ListenAndServe()
}

// stopAPIServer gracefully shutdowns the API server.
func stopAPIServer(ctx context.Context) error {
	if apiServer != nil {
		return apiServer.Server.Shutdown(ctx)
	}

	return nil
}
