// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	hapi "github.com/iliafrenkel/go-pb/src/api/http"
	hweb "github.com/iliafrenkel/go-pb/src/web/http"
)

// Version information, comes from the build flags (see Makefile)
var (
	revision = "unknown"
	version  = "unknown"
	branch   = "unknown"
)

func main() {
	// Set API and Web servers options
	var apiOpts = hapi.APIServerOptions{
		// API server will bind to this address. It follows the same convention
		// as `net.http.Server.Addr`.
		Addr: "127.0.0.1:8000",
		// Maximum body size for any request that accepts body (such as POST).
		MaxBodySize: 10240,
		// Database connection string.
		// It will be a file name for the sqlite database you cab also
		// pass `file::memory:?cache=shared` for in-memory temporary database.
		DBConnectionString: "test.db",
	}
	var webOpts = hweb.WebServerOptions{
		Addr:    "127.0.0.1:8080",
		APIURL:  "http://127.0.0.1:8000",
		Version: version,
	}

	// Create two channels, quit for OS signals and errc for errors comming
	// from the servers.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	errc := make(chan error, 1)

	// Start the API and the Web servers in parallel using Go routines
	log.Println("Starting servers...")
	go func() {
		errc <- startAPIServer(apiOpts)
	}()
	go func() {
		errc <- startWebServer(webOpts)
	}()

	// Wait indefinitely for either one of the OS signals (SIGTERM or SIGINT)
	// or for one of the servers to return an error.
	select {
	case <-quit:
		log.Println("Shutting down servers:")
	case err := <-errc:
		log.Printf("Startup failed, exiting: %v\n", err)
	}

	// If we are here we either received one of the signals or one of the
	// servers encountered an error. Either way, we create a context with
	// timeout to give the servers some time to close all the connections.
	// Please note that the context is shared between the severs. This
	// means that the timeout is for BOTH severs - if the timeout is 10
	// seconds and Web server takes 9 seconds to shutdown it will leave
	// the API server only one second.
	//
	// TODO: #17 Shutdown timeout must be configurable.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := stopWebServer(ctx); err != nil {
		log.Println("\tWeb server forced to shutdown: ", err)
	} else {
		log.Println("\tWeb server is down")
	}

	if err := stopAPIServer(ctx); err != nil {
		log.Println("\tAPI server forced to shutdown: ", err)
	} else {
		log.Println("\tAPI server is down")
	}

	log.Println("Servers are down, sayōnara!")
}
