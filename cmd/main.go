// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	hapi "github.com/iliafrenkel/go-pb/src/api/http"
	hweb "github.com/iliafrenkel/go-pb/src/web/http"
	"github.com/jessevdk/go-flags"
)

// Version information, comes from the build flags (see Makefile)
var (
	revision = "unknown"
	version  = "unknown"
	branch   = "unknown"
)

var opts struct {
	Timeouts struct {
		Shutdown  time.Duration `long:"shutdown" env:"SHUTDOWN" default:"10s" description:"server graceful shutdown timeout"`
		HTTPRead  time.Duration `long:"http-read" env:"HTTP_READ" default:"15s" description:"duration for reading the entire request"`
		HTTPWrite time.Duration `long:"http-write" env:"HTTP_WRITE" default:"15s" description:"duration before timing out writes of the response"`
		HTTPIdle  time.Duration `long:"http-idle" env:"HTTP_IDLE" default:"60s" description:"amount of time to wait for the next request"`
	} `group:"timeout" namespace:"timeout" env-namespace:"GOPB_TIMEOUT"`
	API struct {
		Proto         string `long:"proto" env:"PROTO" default:"http" choice:"http" choice:"https" description:"protocol part of the API server address (http/https)"`
		Host          string `long:"host" env:"HOST" default:"localhost" description:"hostname part of the API server address"`
		Port          uint16 `long:"port" env:"PORT" default:"8000" description:"port part of the API server address"`
		LogFile       string `long:"log-file" env:"LOG_FILE" default:"" description:"full path to the log file, default is stdout"`
		LogMode       string `long:"log-mode" env:"LOG_MODE" default:"production" choice:"debug" choice:"production" description:"log mode, can be 'debug' or 'production'"`
		MaxBodySize   uint   `long:"max-body-size" env:"MAX_BODY_SIZE" default:"10240" description:"maximum size for request's body"`
		DBConnString  string `long:"db-conn-string" env:"DB_CONN_STRING" default:"" description:"full path to the sqlite database file"`
		DBAutoMigrate string `long:"db-auto-migrate" env:"DB_AUTO_MIGRATE" default:"yes" choice:"yes" choice:"no" description:"automatically update the database schema on start up"`
		TokenSecret   string `long:"token-secret" env:"TOKEN_SECRET" default:"" description:"secret string used to sign JWT tokens"`
	} `group:"api" namespace:"api" env-namespace:"GOPB_API"`
	Web struct {
		Proto          string `long:"proto" env:"PROTO" default:"http" choice:"http" choice:"https" description:"protocol part of the Web server address (http/https)"`
		Host           string `long:"host" env:"HOST" default:"localhost" description:"hostname part of the Web server address"`
		Port           uint16 `long:"port" env:"PORT" default:"8080" description:"port part of the Web server address"`
		LogFile        string `long:"log-file" env:"LOG_FILE" default:"" description:"full path to the log file, default is stdout"`
		LogMode        string `long:"log-mode" env:"LOG_MODE" default:"production" choice:"debug" choice:"production" description:"log mode, can be 'debug' or 'production'"`
		CookieAuthKey  string `long:"cookie-auth-key" env:"COOKIE_AUTH_KEY" default:"" description:"secret authentication key, must be 32 or 64 bytes long"`
		BrandName      string `long:"brand-name" env:"BRAND_NAME" default:"Go PB" description:"brand name shown in the header of every page"`
		BrandTagline   string `long:"brand-tagline" env:"BRAND_TAGLINE" default:"A nice and simple pastebin alternative that you can host yourself." description:"brand tagline shown below the brand name"`
		Assets         string `long:"assets" env:"ASSETS" default:"./assets" description:"path to the assets folder"`
		Templates      string `long:"templates" env:"TEMPLATES" default:"./templates" description:"path to the templates folder"`
		BootstrapTheme string `long:"bootstrap-theme" env:"BOOTSTRAP_THEME" default:"original" choice:"flatly" choice:"litera" choice:"materia" choice:"original" choice:"sandstone" choice:"yeti" choice:"zephyr" description:"name of the bootstrap theme to use [flatly, litera, materia, sandstone, yeti or zephyr]"`
		Logo           string `long:"logo" env:"LOGO" default:"" description:"name of the logo image file within the assets folder"`
	} `group:"web" namespace:"web" env-namespace:"GOPB_WEB"`
}

func main() {
	fmt.Printf("go-pb %s\n", version)
	// Parse command line and environment options
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.NamespaceDelimiter = "-"
	p.EnvNamespaceDelimiter = "_"
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			log.Printf("[ERROR] cli error: %v", err)
		}
		os.Exit(2)
	}
	// fmt.Printf("Opts: %+v\n", opts)
	// Set API and Web servers options
	var apiOpts = hapi.APIServerOptions{
		Addr:               opts.API.Host + ":" + fmt.Sprintf("%d", opts.API.Port),
		MaxBodySize:        int64(opts.API.MaxBodySize),
		DBConnectionString: opts.API.DBConnString,
		ReadTimeout:        opts.Timeouts.HTTPRead,
		WriteTimeout:       opts.Timeouts.HTTPWrite,
		IdleTimeout:        opts.Timeouts.HTTPIdle,
		LogFile:            opts.API.LogFile,
		LogMode:            opts.API.LogMode,
		DBAutoMigrate:      opts.API.DBAutoMigrate == "yes",
		TokenSecret:        opts.API.TokenSecret,
	}
	var webOpts = hweb.WebServerOptions{
		Addr:           opts.Web.Host + ":" + fmt.Sprintf("%d", opts.Web.Port),
		Proto:          opts.Web.Proto,
		APIURL:         opts.API.Proto + "://" + opts.API.Host + ":" + fmt.Sprintf("%d", opts.API.Port),
		LogFile:        opts.Web.LogFile,
		LogMode:        opts.Web.LogMode,
		CookieAuthKey:  opts.Web.CookieAuthKey,
		BrandName:      opts.Web.BrandName,
		BrandTagline:   opts.Web.BrandTagline,
		Assets:         opts.Web.Assets,
		Templates:      opts.Web.Templates,
		Logo:           opts.Web.Logo,
		BootstrapTheme: opts.Web.BootstrapTheme,
		Version:        version,
	}

	// Create two channels, quit for OS signals and errc for errors coming
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
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeouts.Shutdown)
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
