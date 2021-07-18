/* Copyright 2021 Ilia Frenkel. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE.txt file.
 */
package main

import (
	"context"
	"log"

	userMem "github.com/iliafrenkel/go-pb/src/api/auth/memory"
	"github.com/iliafrenkel/go-pb/src/api/http"
	pasteMem "github.com/iliafrenkel/go-pb/src/api/paste/memory"
)

var apiServer *http.ApiServer

func StartApiServer(opts http.ApiServerOptions) error {
	apiServer = http.New(pasteMem.New(), userMem.New(), opts)

	log.Println("API server listening on ", opts.Addr)

	return apiServer.ListenAndServe()
}

func StopApiServer(ctx context.Context) error {
	if apiServer != nil {
		return apiServer.Server.Shutdown(ctx)
	}

	return nil
}
