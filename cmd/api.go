/* Copyright 2021 Ilia Frenkel. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE.txt file.
 */
package main

import (
	"context"
	"fmt"
	"log"

	u "github.com/iliafrenkel/go-pb/src/api/auth/sqlite"
	"github.com/iliafrenkel/go-pb/src/api/http"
	p "github.com/iliafrenkel/go-pb/src/api/paste/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var apiServer *http.ApiServer
var db *gorm.DB

func StartApiServer(opts http.ApiServerOptions) error {
	// Connect to the database
	var err error
	db, err = gorm.Open(sqlite.Open(opts.DBConnectionString), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to establish database connection: %w", err)
	}

	// Create UserService
	userSvc, err := u.New(u.SvcOptions{DBConnection: db})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to create UserService: %w", err)
	}

	// Create PasteService
	pasteSvc, err := p.New(p.SvcOptions{DBConnection: db})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to create PasteService: %w", err)
	}

	apiServer = http.New(pasteSvc, userSvc, opts)

	log.Println("API server listening on ", opts.Addr)

	return apiServer.ListenAndServe()
}

func StopApiServer(ctx context.Context) error {
	if apiServer != nil {
		return apiServer.Server.Shutdown(ctx)
	}

	return nil
}
