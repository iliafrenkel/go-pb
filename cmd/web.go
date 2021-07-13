package main

import (
	"context"
	"log"

	"github.com/iliafrenkel/go-pb/src/web/http"
)

var webServer *http.WebServer

func StartWebServer() error {
	addr := "127.0.0.1:8080"
	webServer := http.New(http.WebServerOptions{ApiURL: "http://127.0.0.1:8000"})

	log.Println("Web server listening on ", addr)

	return webServer.ListenAndServe(addr)
}

func StopWebServer(ctx context.Context) error {
	if webServer != nil {
		return webServer.Server.Shutdown(ctx)
	}

	return nil
}
