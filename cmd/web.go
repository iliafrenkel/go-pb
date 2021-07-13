package main

import (
	"context"
	"log"

	"github.com/iliafrenkel/go-pb/src/web/http"
)

var webServer *http.WebServer

func StartWebServer(opts http.WebServerOptions) error {
	webServer := http.New(opts)

	log.Println("Web server listening on ", opts.Addr)

	return webServer.ListenAndServe()
}

func StopWebServer(ctx context.Context) error {
	if webServer != nil {
		return webServer.Server.Shutdown(ctx)
	}

	return nil
}
