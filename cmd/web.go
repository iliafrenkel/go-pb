// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.
package main

import (
	"context"
	"log"

	"github.com/iliafrenkel/go-pb/src/web/http"
)

var webServer *http.WebServer

// startWebServer initialises and starts the WebServer.
func startWebServer(opts http.WebServerOptions) error {
	webServer := http.New(opts)

	log.Println("Web server listening on ", opts.Addr)

	return webServer.ListenAndServe()
}

// stopWebServer gracefully shuts down the WebServer.
func stopWebServer(ctx context.Context) error {
	if webServer != nil {
		return webServer.Server.Shutdown(ctx)
	}

	return nil
}
