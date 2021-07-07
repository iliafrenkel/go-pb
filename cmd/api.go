package main

import (
	"github.com/iliafrenkel/go-pb/src/api/db/memory"
	"github.com/iliafrenkel/go-pb/src/api/http"
)

func StartApiServer() error {
	var api *http.ApiServer = http.New(memory.New())

	return api.ListenAndServe(":8080")
}
