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
)

var apiServer *http.ApiServer

func StartApiServer(opts http.ApiServerOptions) error {
	userSvc, err := u.New(u.DBOptions{Connection: opts.DBConnection})
	if err != nil {
		return fmt.Errorf("StartApiServer: failed to create UserService: %w", err)
	}

	pasteSvc, err := p.New(p.DBOptions{Connection: opts.DBConnection})
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
