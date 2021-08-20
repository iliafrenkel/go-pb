// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pkgz/lgr"
	"github.com/iliafrenkel/go-pb/src/web"
	"github.com/jessevdk/go-flags"
)

// Version information, comes from the build flags (see Makefile)
var (
	version = `¯\_(ツ)_/¯`
)

var opts struct {
	Timeouts struct {
		Shutdown  time.Duration `long:"shutdown" env:"SHUTDOWN" default:"10s" description:"server graceful shutdown timeout"`
		HTTPRead  time.Duration `long:"http-read" env:"HTTP_READ" default:"15s" description:"duration for reading the entire request"`
		HTTPWrite time.Duration `long:"http-write" env:"HTTP_WRITE" default:"15s" description:"duration before timing out writes of the response"`
		HTTPIdle  time.Duration `long:"http-idle" env:"HTTP_IDLE" default:"60s" description:"amount of time to wait for the next request"`
	} `group:"timeout" namespace:"timeout" env-namespace:"GOPB_TIMEOUT"`
	Web struct {
		Proto          string `long:"proto" env:"PROTO" default:"http" choice:"http" choice:"https" description:"protocol part of the Web server address (http/https)"`
		Host           string `long:"host" env:"HOST" default:"localhost" description:"hostname part of the Web server address"`
		Port           uint16 `long:"port" env:"PORT" default:"8080" description:"port part of the Web server address"`
		LogFile        string `long:"log-file" env:"LOG_FILE" default:"" description:"full path to the log file, default is stdout"`
		LogMode        string `long:"log-mode" env:"LOG_MODE" default:"production" choice:"debug" choice:"production" description:"log mode, can be 'debug' or 'production'"`
		BrandName      string `long:"brand-name" env:"BRAND_NAME" default:"Go PB" description:"brand name shown in the header of every page"`
		BrandTagline   string `long:"brand-tagline" env:"BRAND_TAGLINE" default:"A nice and simple pastebin alternative that you can host yourself." description:"brand tagline shown below the brand name"`
		Assets         string `long:"assets" env:"ASSETS" default:"./assets" description:"path to the assets folder"`
		Templates      string `long:"templates" env:"TEMPLATES" default:"./templates" description:"path to the templates folder"`
		BootstrapTheme string `long:"bootstrap-theme" env:"BOOTSTRAP_THEME" default:"original" choice:"flatly" choice:"litera" choice:"materia" choice:"original" choice:"sandstone" choice:"yeti" choice:"zephyr" description:"name of the bootstrap theme to use [flatly, litera, materia, sandstone, yeti or zephyr]"`
		Logo           string `long:"logo" env:"LOGO" default:"bighead.svg" description:"name of the logo image file within the assets folder"`
		MaxBodySize    int64  `long:"max-body-size" env:"MAX_BODY_SIZE" default:"10240" description:"maximum size for request's body"`
	} `group:"web" namespace:"web" env-namespace:"GOPB_WEB"`
	DB struct {
		Type       string `long:"type" env:"TYPE" default:"memory" choice:"memory" choice:"postgres" description:"database type to use for storage"`
		Connection string `long:"connection" env:"CONNECTION" default:"" description:"database connection string, ignored for memory"`
	} `group:"db" namespace:"db" env-namespace:"GOPB_DB"`
	Auth struct {
		Secret         string        `long:"secret" env:"SECRET" default:"" description:"secret used for JWT token generation/verification"`
		TokenDuration  time.Duration `long:"token-duration" env:"TOKEN_DURATION" default:"5m" description:"JWT token expiration"`
		CookieDuration time.Duration `long:"cookie-duration" env:"COOKIE_DURATION" default:"24h" description:"cookie expiration"`
		Issuer         string        `long:"issuer" env:"ISSUER" default:"go-pb" description:"app name used to oauth requests"`
		URL            string        `long:"url" env:"URL" default:"http://localhost:8080" description:"callback url for oauth requests"`
		GitHubCID      string        `long:"github-cid" env:"GITHUB_CID" default:"" description:"github client id used for oauth"`
		GitHubCSEC     string        `long:"github-csec" env:"GITHUB_CSEC" default:"" description:"github client secret used for oauth"`
		GoogleCID      string        `long:"google-cid" env:"GOOGLE_CID" default:"" description:"google client id used for oauth"`
		GoogleCSEC     string        `long:"google-csec" env:"GOOGLE_CSEC" default:"" description:"google client secret used for oauth"`
		TwitterCID     string        `long:"twitter-cid" env:"TWITTER_CID" default:"" description:"twitter client id used for oauth"`
		TwitterCSEC    string        `long:"twitter-csec" env:"TWITTER_CSEC" default:"" description:"twitter client secret used for oauth"`
	} `group:"auth" namespace:"auth" env-namespace:"GOPB_AUTH"`
	Debug bool `long:"debug" env:"DEBUG" description:"debug mode"`
}

func main() {
	// Say hello
	fmt.Printf("go-pb %s\n", version)

	// Parse the flags
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	p.NamespaceDelimiter = "-"
	p.EnvNamespaceDelimiter = "_"
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("[ERROR] cli error: %v", err)
		}
		os.Exit(2)
	}

	log := setupLog(opts.Debug)

	if opts.Debug {
		log.Logf("INFO Options: %+v", opts)
	}

	// Start the server
	webServer := web.New(log, web.WebServerOptions{
		Addr:               opts.Web.Host + ":" + fmt.Sprintf("%d", opts.Web.Port),
		Proto:              opts.Web.Proto,
		ReadTimeout:        opts.Timeouts.HTTPRead,
		WriteTimeout:       opts.Timeouts.HTTPWrite,
		IdleTimeout:        opts.Timeouts.HTTPIdle,
		LogFile:            opts.Web.LogFile,
		LogMode:            opts.Web.LogMode,
		BrandName:          opts.Web.BrandName,
		BrandTagline:       opts.Web.BrandTagline,
		Assets:             opts.Web.Assets,
		Templates:          opts.Web.Templates,
		Logo:               opts.Web.Logo,
		MaxBodySize:        opts.Web.MaxBodySize,
		BootstrapTheme:     opts.Web.BootstrapTheme,
		Version:            version,
		AuthSecret:         opts.Auth.Secret,
		AuthTokenDuration:  opts.Auth.TokenDuration,
		AuthCookieDuration: opts.Auth.CookieDuration,
		AuthIssuer:         opts.Auth.Issuer,
		AuthURL:            opts.Auth.URL,
		DBType:             opts.DB.Type,
		DBConn:             opts.DB.Connection,
		GitHubCID:          opts.Auth.GitHubCID,
		GitHubCSEC:         opts.Auth.GitHubCSEC,
		GoogleCID:          opts.Auth.GoogleCID,
		GoogleCSEC:         opts.Auth.GoogleCSEC,
		TwitterCID:         opts.Auth.TwitterCID,
		TwitterCSEC:        opts.Auth.TwitterCSEC,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	errc := make(chan error, 1)

	go func() {
		log.Logf("INFO Web server listening on %s:%d", opts.Web.Host, opts.Web.Port)
		errc <- webServer.ListenAndServe()
	}()

	// Wait indefinitely for either one of the OS signals (SIGTERM or SIGINT)
	// or for one of the servers to return an error.
	select {
	case <-quit:
		log.Logf("INFO Shutting down ...")
	case err := <-errc:
		log.Logf("ERROR Startup failed, exiting: %v\n", err)
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

	if err := webServer.Shutdown(ctx); err != nil {
		log.Logf("INFO \tWeb server forced to shutdown: %v\n", err)
	} else {
		log.Logf("INFO \tWeb server is down")
	}
	log.Logf("INFO Sayōnara!")
}

func setupLog(dbg bool) *lgr.Logger {
	if dbg {
		return lgr.New(lgr.Debug, lgr.CallerFile, lgr.CallerFunc, lgr.Msec, lgr.LevelBraces)
	}
	return lgr.New()
}
