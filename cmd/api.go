package main

import (
	"context"

	"github.com/iliafrenkel/go-pb/src/api/db/memory"
	"github.com/iliafrenkel/go-pb/src/api/http"
)

var apiServer *http.ApiServer

func StartApiServer() error {
	apiServer = http.New(memory.New())

	return apiServer.ListenAndServe(":8080")
}

func StopApiServer(ctx context.Context) error {
	if apiServer != nil {
		return apiServer.Server.Shutdown(ctx)
	}

	return nil
}
