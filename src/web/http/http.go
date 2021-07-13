package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/src/api"
)

// WebServerOptions defines various parameters needed to run the WebServer
type WebServerOptions struct {
	// ApiURL specifies the full URL of the ApiServer withouth the trailing
	// backslash such as "http://localhost:8000".
	ApiURL string
}

// WebServer encapsulates a router and a server.
// Normally, you'd create a new instance by calling New which configures the
// rotuer and then call ListenAndServe to start serving incoming requests.
type WebServer struct {
	Router  *gin.Engine
	Server  *http.Server
	Options WebServerOptions
}

// New returns an instance of the WebServer with initialised middleware,
// loaded templates and routes. You can call ListenAndServe on a newly
// created instance to initialise the HTTP server and start handling incoming
// requests.
func New(opts WebServerOptions) *WebServer {
	var handler WebServer
	handler.Options = opts

	// Initialise the router and load the templates from /src/web/templates folder.
	handler.Router = gin.Default()
	handler.Router.LoadHTMLGlob(filepath.Join("..", "src", "web", "templates", "*.html"))
	handler.Router.Static("/assets", "../src/web/assets")

	// Define all the routes
	handler.Router.GET("/", handler.handleRoot)
	handler.Router.GET("/ping", handler.handlePing)
	handler.Router.GET("/u/login", handler.handleUserLogin)
	handler.Router.GET("/u/register", handler.handleUserRegister)
	handler.Router.GET("/p/:id", handler.handlePaste)
	handler.Router.POST("/p/", handler.handlePasteCreate)

	// Catch all route just shows the 404 error page
	handler.Router.NoRoute(func(c *gin.Context) {
		c.Set("errorCode", http.StatusNotFound)
		c.Set("errorText", http.StatusText(http.StatusNotFound))
		c.Set("errorMessage", "Unfortunately the page you are looking for is not there 🙁")
		handler.showError(c)
	})

	return &handler
}

// ListenAndServe starts an HTTP server and binds it to the provided address.
//
// TODO: Timeouts should be configurable.
func (h *WebServer) ListenAndServe(addr string) error {
	h.Server = &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h.Router,
	}

	return h.Server.ListenAndServe()
}

// handleRoot returns the home page and should be bound to the '/' URL.
// It assumes that there is a template named "index.html" and that it
// was already loaded.
func (h *WebServer) handleRoot(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"title": "Go PB - Home",
		},
	)
}

// handlePing returns a simple JSON object: {"message":"pong"}. It is usually
// set to handle GET /ping route and is used as a healthcheck.
func (h *WebServer) handlePing(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// handleUserLogin returns a page with the login form. It assumes that there is
// a template named "login.html" and that it was already loaded.
func (h *WebServer) handleUserLogin(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"login.html",
		gin.H{
			"title": "Go PB - Login",
		},
	)
}

// handleUserRegister returns a page with the registration form. It assumes
// that there is a template named "register.html" and that it was already
// loaded.
func (h *WebServer) handleUserRegister(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"register.html",
		gin.H{
			"title": "Go PB - Register",
		},
	)
}

// handlePaste queries the API for a paste and returns a page that displays it.
func (h *WebServer) handlePaste(c *gin.Context) {
	// Query the API for a paste by ID
	id := c.Param("id")
	resp, err := http.Get(h.Options.ApiURL + "/paste/" + id)
	// API server maybe down or some other network error
	if err != nil {
		log.Println("handlePaste: error querying API: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		h.showError(c)
		return
	}
	// If API response code is not 200 return it and log an error
	if resp.StatusCode != http.StatusOK {
		log.Println("handlePaste: API returned: ", resp.StatusCode)
		c.Set("errorCode", resp.StatusCode)
		c.Set("errorText", http.StatusText(resp.StatusCode))
		if resp.StatusCode == http.StatusNotFound {
			c.Set("errorMessage", "The paste cannot be found.")
		}
		h.showError(c)
		return
	}
	// Read response body and try to parse it as JSON
	var p api.Paste
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("handlePaste: failed to read API response body: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	if err := json.Unmarshal(b, &p); err != nil {
		log.Println("handlePaste: failed to parse API response", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	// Send HTML
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"Title":    p.Title,
			"Body":     p.Body,
			"Language": p.Syntax,
			"URL":      p.URL(),
			"Server":   "http://localhost:8080", //TODO: this has to come from somewhere
		},
	)
}

// handlePasteCreate collects information from the new paste form and calls
// the API to create a new paste. If successful it shows the new paste.
func (h *WebServer) handlePasteCreate(c *gin.Context) {
	var p api.Paste
	// Try to parse the form
	if err := c.ShouldBind(&p); err != nil {
		log.Println("handlePasteCreate: failed to bind to form data: ", err)
		c.Set("errorCode", http.StatusBadRequest)
		c.Set("errorText", http.StatusText(http.StatusBadRequest))
		h.showError(c)
		return
	}
	// Try to create a new paste by calling the API
	paste, _ := json.Marshal(p)
	resp, err := http.Post(h.Options.ApiURL+"/paste", "application/json", bytes.NewBuffer(paste)) // TODO: API address must come from configuration

	if err != nil {
		log.Println("handlePasteCreate: error talking to API: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	// Check API response status
	if resp.StatusCode != http.StatusCreated {
		log.Println("handlePasteCreate: API returned: ", resp.StatusCode)
		c.Set("errorCode", resp.StatusCode)
		c.Set("errorText", http.StatusText(resp.StatusCode))
		h.showError(c)
		return
	}
	// Get API response body and try to parse it as JSON
	var data api.Paste
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("handlePasteCreate: failed to read API response body: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Println("handlePasteCreate: failed to parse API response", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	// Send back HTML that display newly created paste
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"Title":    data.Title,
			"Body":     data.Body,
			"Language": data.Syntax,
			"URL":      resp.Header.Get("Location"),
			"Server":   "http://localhost:8080", //TODO: this has to come from somewhere
		},
	)
}

// showError displays a custom error page using error.html template.
// The context can use "errorCode", "errorText" and "errorMessage" keys to
// customise what is shown on the page.
func (h *WebServer) showError(c *gin.Context) {
	var (
		errorCode int
		errorText string
		errorMsg  string
	)
	if val, ok := c.Get("errorCode"); ok {
		errorCode = val.(int)
	} else {
		errorCode = http.StatusNotImplemented
	}
	if val, ok := c.Get("errorText"); ok {
		errorText = val.(string)
	} else {
		errorText = http.StatusText(http.StatusNotImplemented)
	}
	if val, ok := c.Get("errorMessage"); ok {
		errorMsg = val.(string)
	}

	c.HTML(
		errorCode,
		"error.html",
		gin.H{
			"title":        "Error",
			"errorCode":    errorCode,
			"errorText":    errorText,
			"errorMessage": errorMsg,
		},
	)
}
