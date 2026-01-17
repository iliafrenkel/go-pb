// Package web implements a web server that provides a front-end for the
// go-pb application.
package web

import (
	"context"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-pkgz/auth/v2"
	"github.com/go-pkgz/auth/v2/avatar"
	"github.com/go-pkgz/auth/v2/token"
	"github.com/go-pkgz/lgr"
	"github.com/iliafrenkel/go-pb/src/service"
	"github.com/iliafrenkel/go-pb/src/store"
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
	store.DiskConfig
}

// Server encapsulates a router and a server.
// Normally, you'd create a new instance by calling New which configures the
// rotuer and then call ListenAndServe to start serving incoming requests.
type Server struct {
	router    *router
	server    *http.Server
	options   ServerOptions
	templates *template.Template
	log       *lgr.Logger
	service   *service.Service
}

// router is a wrapper around http.ServeMux that allows to set a NotFoundHandler
type router struct {
	*http.ServeMux
	notFoundHandler http.Handler
}

// ServeHTTP makes the router implement the http.Handler interface.
func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h, _ := r.Handler(req)
	if h == nil {
		r.notFoundHandler.ServeHTTP(w, req)
		return
	}
	h.ServeHTTP(w, req)
}

// NotFound sets the not-found handler.
func (r *router) NotFound(h http.Handler) {
	r.notFoundHandler = h
}

// ListenAndServe starts an HTTP server and binds it to the provided address.
// You have to call New() first to initialise the WebServer.
func (h *Server) ListenAndServe() error {
	h.server = &http.Server{
		Addr:         h.options.Addr,
		WriteTimeout: h.options.WriteTimeout,
		ReadTimeout:  h.options.ReadTimeout,
		IdleTimeout:  h.options.IdleTimeout,
		Handler:      h.router,
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
	case "disk":
		handler.service, err = service.NewWithDiskDB(&opts.DiskConfig)
		if err != nil {
			handler.log.Logf("FATAL error creating Disk storage service: %v", err)
		}
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
	mux := http.NewServeMux()
	handler.router = &router{
		ServeMux:        mux,
		notFoundHandler: http.HandlerFunc(handler.notFound),
	}
	// Templates and static files
	handler.router.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(handler.options.Assets))))

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

	if opts.LogMode == "debug" {
		authSvc.AddProvider("dev", "", "") // dev auth, runs dev oauth2 server on :8084

		go func() {
			devAuthServer, err := authSvc.DevAuth()
			if err != nil {
				handler.log.Logf("FATAL %v", err)
			}

			devAuthServer.Run(context.Background())
		}()
	}

	m := authSvc.Middleware()
	authRoutes, avaRoutes := authSvc.Handlers()
	handler.router.Handle("/auth/", authRoutes)
	handler.router.Handle("/avatar/", avaRoutes)

	// Define routes
	pHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/p/")
		if r.Method == http.MethodPost {
			if len(id) > 0 {
				handler.handleGetPastePage(w, r) // This is for password-protected pastes
				return
			}
			handler.handlePostPaste(w, r) // This is for new pastes
			return
		}
		if r.Method == http.MethodGet {
			if len(id) > 0 {
				handler.handleGetPastePage(w, r)
				return
			}
			handler.handleGetHomePage(w, r)
			return
		}
		handler.notFound(w, r)
	})
	handler.router.Handle("/p/", m.Trace(pHandler))
	handler.router.Handle("/l/", m.Trace(http.HandlerFunc(handler.handleGetPastesList)))
	handler.router.Handle("/a/", m.Trace(http.HandlerFunc(handler.handleGetArchive)))
	// The default handler will catch everything that is not handled by other handlers.
	handler.router.Handle("/", m.Trace(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We only want to handle the root path here.
		// Everything else should be a 404.
		if r.URL.Path != "/" {
			handler.notFound(w, r)
			return
		}
		handler.handleGetHomePage(w, r)
	})))

	return &handler
}
