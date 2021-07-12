package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/src/api"
)

type WebServer struct {
	Router *gin.Engine
	Server *http.Server
}

// New returns an instance of the WebServer with initialised middleware,
// loaded templates and routes. You can call ListenAndServe on a newly
// created instance to initialise the HTTP server and start handling incoming
// requests.
func New() *WebServer {
	var handler WebServer

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
			"title": " Go PB - Home",
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
			"title": " Go PB - Login",
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
			"title": " Go PB - Register",
		},
	)
}

// handlePaste queries the API for a paste and returns a page that displays it.
func (h *WebServer) handlePaste(c *gin.Context) {
	// Query the API for a paste by ID
	id := c.Param("id")
	resp, err := http.Get("http://localhost:8000/paste/" + id) // TODO: API address has to come from configuration
	// API server maybe down or some other network error
	if err != nil {
		log.Println("handlePaste: error querying API: ", err)
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	// If API response code is not 200 return it and log an error
	if resp.StatusCode != http.StatusOK {
		log.Println("handlePaste: API returned: ", resp.StatusCode)
		c.String(resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}
	// Read response body and try to parse it as JSON
	var p api.Paste
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("handlePaste: failed to read API response body: ", err)
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if err := json.Unmarshal(b, &p); err != nil {
		log.Println("handlePaste: failed to parse API response", err)
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
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

func (h *WebServer) handlePasteCreate(c *gin.Context) {
	var p api.Paste
	var data api.Paste
	// Get the paste title and body from the form
	if b, ok := c.GetPostForm("body"); !ok || len(b) == 0 {
		c.String(http.StatusBadRequest, "body cannot be empty")
		return
	}
	p.Body = c.PostForm("body")
	p.Title = c.DefaultPostForm("title", "untitled")
	p.DeleteAfterRead, _ = strconv.ParseBool(c.PostForm("delete_after_read"))
	p.Syntax = c.DefaultPostForm("syntax", "none")

	// Try to create a new paste by calling the API
	paste, _ := json.Marshal(p)
	resp, err := http.Post("http://localhost:8000/paste", "application/json", bytes.NewBuffer(paste))

	if err != nil {
		log.Println("handlePasteCreate: error talking to API: ", err)
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	// Check API response status
	if resp.StatusCode != http.StatusCreated {
		log.Println("handlePasteCreate: API returned: ", resp.StatusCode)
		c.String(resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}

	// Get API response body and try to parse it as JSON
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("handlePasteCreate: failed to read API response body: ", err)
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Println("handlePasteCreate: failed to parse API response", err)
		c.String(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
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