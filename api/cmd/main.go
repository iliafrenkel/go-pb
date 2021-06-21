package main

import (
	"os"

	"github.com/iliafrenkel/go-pb/api/http"
	"github.com/iliafrenkel/go-pb/api/memory"
)

func main() {
	var api *http.ApiHandler = http.New(memory.New())

	api.ListenAndServe(":8080")

	os.Exit(0)
}
