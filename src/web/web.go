// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package web implements a web server that provides a front-end for the
// go-pb application.
package web

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/lgr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/iliafrenkel/go-pb/src/service"
)

// ServerOptions defines various parameters needed to run the WebServer
type ServerOptions struct {
	Addr               string        // address to listen on, see http.Server docs for details
	Proto              string        // protocol, either "http" or "https"
	ReadTimeout        time.Duration // maximum duration for reading the entire request.
	WriteTimeout       time.Duration // maximum duration before timing out writes of the response
	IdleTimeout        time.Duration // maximum amount of time to wait for the next request
	LogFile            string        // if not empty, will write logs to the file
	LogMode            string        // can be either "debug" or "production"
	BrandName          string        // displayed at the top of each page, default is "Go PB"
	BrandTagline       string        // displayed below the BrandName
	Assets             string        // location of the assets folder (css, js, images)
	Templates          string        // location of the templates folder
	Logo               string        // name of the logo image within the assets folder
	MaxBodySize        int64         // maximum size for request's body
	BootstrapTheme     string        // one of the themes, see css files in the assets folder
	Version            string        // app version, comes from build
	AuthSecret         string        // secret for JWT token generation and validation
	AuthTokenDuration  time.Duration // JWT token expiration duration
	AuthCookieDuration time.Duration // cookie expiration time
	AuthIssuer         string        // application name used as an issuer in oauth requests
	AuthURL            string        // callback URL for oauth requests
	DBType             string        // type of the store to use
	DBConn             string        // database connection string
	GitHubCID          string        // github client id for oauth
	GitHubCSEC         string        // github client secret for oauth
	GoogleCID          string        // google client id for oauth
	GoogleCSEC         string        // google client secret for oauth
	TwitterCID         string        // twitter client id for oauth
	TwitterCSEC        string        // twitter client secret for oauth
}

// Server encapsulates a router and a server.
// Normally, you'd create a new instance by calling New which configures the
// rotuer and then call ListenAndServe to start serving incoming requests.
type Server struct {
	router    *mux.Router
	server    *http.Server
	options   ServerOptions
	templates *template.Template
	log       *lgr.Logger
	service   *service.Service
}

var dbgLogFormatter handlers.LogFormatter = func(writer io.Writer, params handlers.LogFormatterParams) {
	const (
		green   = "\033[97;42m"
		white   = "\033[90;47m"
		yellow  = "\033[90;43m"
		red     = "\033[97;41m"
		blue    = "\033[97;44m"
		magenta = "\033[97;45m"
		cyan    = "\033[97;46m"
		reset   = "\033[0m"
	)

	code := params.StatusCode
	cclr := ""
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		cclr = green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		cclr = white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		cclr = yellow
	default:
		cclr = red
	}

	method := params.Request.Method
	mclr := ""
	switch method {
	case http.MethodGet:
		mclr = blue
	case http.MethodPost:
		mclr = cyan
	case http.MethodPut:
		mclr = yellow
	case http.MethodDelete:
		mclr = red
	case http.MethodPatch:
		mclr = green
	case http.MethodHead:
		mclr = magenta
	case http.MethodOptions:
		mclr = white
	default:
		mclr = reset
	}

	host, _, err := net.SplitHostPort(params.Request.RemoteAddr)
	if err != nil {
		host = params.Request.RemoteAddr
	}

	fmt.Fprintf(writer, "|%s %3d %s| %15s |%s %-7s %s| %8d | %s \n",
		cclr, code, reset,
		host,
		mclr, method, reset,
		params.Size,
		params.URL.RequestURI(),
	)
}

// ListenAndServe starts an HTTP server and binds it to the provided address.
// You have to call New() first to initialise the WebServer.
func (h *Server) ListenAndServe() error {
	var hdlr http.Handler
	var w io.Writer
	var err error
	if h.options.LogFile == "" {
		w = lgr.ToWriter(h.log, "")
	} else {
		w, err = os.OpenFile(h.options.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("WebServer.ListenAndServer: cannot open log file: [%s]: %w", h.options.LogFile, err)
		}
	}
	if h.options.LogMode == "debug" {
		hdlr = handlers.CustomLoggingHandler(w, h.router, dbgLogFormatter)
	} else {
		hdlr = handlers.CombinedLoggingHandler(w, h.router)
	}
	h.server = &http.Server{
		Addr:         h.options.Addr,
		WriteTimeout: h.options.WriteTimeout,
		ReadTimeout:  h.options.ReadTimeout,
		IdleTimeout:  h.options.IdleTimeout,
		Handler:      hdlr,
	}

	return h.server.ListenAndServe()
}

// Shutdown gracefully shutdown the server with the givem context.
func (h *Server) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

// New returns an instance of the WebServer with initialised middleware,
// loaded templates and routes. You can call ListenAndServe on a newly
// created instance to initialise the HTTP server and start handling incoming
// requests.
func New(l *lgr.Logger, opts ServerOptions) *Server {
	var handler Server
	handler.log = l
	handler.options = opts

	// Load template
	tpl, err := template.ParseGlob(handler.options.Templates + "/*.html")
	if err != nil {
		handler.log.Logf("FATAL error loading templates: %v", err)
	}
	handler.log.Logf("INFO loaded %d templates", len(tpl.Templates()))
	handler.templates = tpl

	// Initialise the service
	switch opts.DBType {
	case "memory":
		handler.service = service.NewWithMemDB()
	case "postgres":
		handler.service, err = service.NewWithPostgres(opts.DBConn)
		if err != nil {
			handler.log.Logf("FATAL error creating Postgres service: %v", err)
		}
	default:
		handler.log.Logf("FATAL unknown store type: %v", opts.DBType)
	}

	// Initialise the router
	handler.router = mux.NewRouter()

	// Templates and static files
	handler.router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(handler.options.Assets))))

	// Auth middleware
	authSvc := auth.NewService(auth.Opts{
		SecretReader: token.SecretFunc(func(id string) (string, error) { // secret key for JWT
			return handler.options.AuthSecret, nil
		}),
		TokenDuration:  handler.options.AuthTokenDuration,
		CookieDuration: handler.options.AuthCookieDuration,
		Issuer:         handler.options.AuthIssuer,
		URL:            handler.options.AuthURL,
		DisableXSRF:    true,
		AvatarStore:    avatar.NewLocalFS(".tmp"),
		Logger:         handler.log, // optional logger for auth library
	})
	authSvc.AddProvider("github", handler.options.GitHubCID, handler.options.GitHubCSEC)
	authSvc.AddProvider("google", handler.options.GoogleCID, handler.options.GoogleCSEC)
	authSvc.AddProvider("twitter", handler.options.TwitterCID, handler.options.TwitterCSEC)
	authSvc.AddProvider("dev", "", "") // dev auth, runs dev oauth2 server on :8084

	go func() {
		devAuthServer, err := authSvc.DevAuth()
		if err != nil {
			handler.log.Logf("FATAL %v", err)
		}
		devAuthServer.Run(context.Background())
	}()
	m := authSvc.Middleware()
	handler.router.Use(m.Trace)
	authRoutes, avaRoutes := authSvc.Handlers()
	handler.router.PathPrefix("/auth").Handler(authRoutes)
	handler.router.PathPrefix("/avatar").Handler(avaRoutes)

	// Define routes
	handler.router.HandleFunc("/", handler.handleGetHomePage).Methods("GET")
	handler.router.HandleFunc("/p/", handler.handlePostPaste).Methods("POST")
	handler.router.HandleFunc("/p/", handler.handleGetHomePage).Methods("GET")
	handler.router.HandleFunc("/p/{id}", handler.handleGetPastePage).Methods("GET")
	handler.router.HandleFunc("/p/{id}", handler.handleGetPastePage).Methods("POST")
	handler.router.HandleFunc("/l/", handler.handleGetPastesList).Methods("GET")
	handler.router.HandleFunc("/a/", handler.handleGetArchive).Methods("GET")

	// Common error routes
	handler.router.NotFoundHandler = handler.router.NewRoute().BuildOnly().HandlerFunc(handler.notFound).GetHandler()

	return &handler
}
