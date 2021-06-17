package main

import (
	"log"

	"github.com/iliafrenkel/go-pb/api/http"
	"github.com/iliafrenkel/go-pb/api/memory"
)

func main() {
	var api *http.ApiHandler = http.New(memory.New())

	log.Fatal(api.ListenAndServe(":8080"))
}
