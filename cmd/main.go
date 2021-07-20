/* Copyright 2021 Ilia Frenkel. All rights reserved.
 * Use of this source code is governed by a MIT-style
 * license that can be found in the LICENSE.txt file.
 */
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
	var apiOpts = hapi.ApiServerOptions{
		// API server will bind to this address. It follows the same convention
		// as `net.http.Server.Addr`.
		Addr: "127.0.0.1:8000",
		// Maximum body size for any request that accepts body (such as POST).
		MaxBodySize: 10240,
		// Database connection string.
		// It will be a file name for the sqlite database you cab also
		// pass `file::memory:?cache=shared` for in-memory temporary database.
		DBConnection: "test.db",
	}
	var webOpts = hweb.WebServerOptions{
		Addr:    "127.0.0.1:8080",
		ApiURL:  "http://127.0.0.1:8000",
		Version: version,
	}

	errc := make(chan error, 1)

	log.Println("Starting servers...")
	// We start the Web server after the API one so that no web requests
	// come before API is ready. The shutdown is done in reverse order.
	go func() {
		errc <- StartApiServer(apiOpts)
	}()
	go func() {
		errc <- StartWebServer(webOpts)
	}()

	// Graceful shutdown - we create a channel for system signals and
	// "subscribe" to SIGINT or SIGTERM. We then wait indefinitely for
	// one of the signals.
	// Once we receive a signal we create a context with timeout to give
	// the servers some time to close all the connections. Please note
	// that the context is shared between the severs. This means that the
	// timeout is for BOTH severs - if the timeout is 10 seconds and web
	// server takes 9 seconds to shutdown it will leave the API server
	// only one second.
	//
	// TODO: Shutdown timeout must be configurable.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Shutting down servers:")
	case err := <-errc:
		log.Printf("Startup failed: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := StopWebServer(ctx); err != nil {
		log.Println("\tWeb server forced to shutdown: ", err)
	} else {
		log.Println("\tWeb server is down")
	}

	if err := StopApiServer(ctx); err != nil {
		log.Println("\tAPI server forced to shutdown: ", err)
	} else {
		log.Println("\tAPI server is down")
	}

	log.Println("Servers are down, sayōnara!")
}
