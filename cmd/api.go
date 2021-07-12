package main

import (
	"context"
	"log"

	"github.com/iliafrenkel/go-pb/src/api/db/memory"
	"github.com/iliafrenkel/go-pb/src/api/http"
)

var apiServer *http.ApiServer

func StartApiServer() error {
	addr := "127.0.0.1:8000"
	apiServer = http.New(memory.New())

	log.Println("API server listening on ", addr)

	return apiServer.ListenAndServe(addr)
}

func StopApiServer(ctx context.Context) error {
	if apiServer != nil {
		return apiServer.Server.Shutdown(ctx)
	}

	return nil
}
