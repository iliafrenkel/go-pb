package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"unicode"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/auth"
)

// WebServerOptions defines various parameters needed to run the WebServer
type WebServerOptions struct {
	// Addr will be passed to http.Server to listen on, see http.Server
	// documentation for more information.
	Addr string
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

	// Sessions management - setup sessions with the cookie store.
	store := cookie.NewStore([]byte("hardcodedsecret")) //TODO: move the secret to env
	store.Options(sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   0,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	handler.Router.Use(sessions.Sessions("gopb", store))
	handler.Router.Use(handler.handleSession)

	// Templates and static files
	handler.Router.LoadHTMLGlob(filepath.Join("..", "src", "web", "templates", "*.html"))
	handler.Router.Static("/assets", "../src/web/assets")

	// Define all the routes
	handler.Router.GET("/", handler.handleRoot)
	handler.Router.GET("/ping", handler.handlePing)
	handler.Router.GET("/u/login", handler.handleUserLogin)
	handler.Router.POST("/u/login", handler.handleDoUserLogin)
	handler.Router.GET("/u/logout", handler.handleDoUserLogout)
	handler.Router.GET("/u/register", handler.handleUserRegister)
	handler.Router.POST("/u/register", handler.handleDoUserRegister)
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
// You have to call New() first to initialise the WebServer.
//
// TODO: Timeouts should be configurable.
func (h *WebServer) ListenAndServe() error {
	h.Server = &http.Server{
		Addr:         h.Options.Addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h.Router,
	}

	return h.Server.ListenAndServe()
}

// handleSession validates JWT token and updates the session accordingly and
// it does so for every request.
//
// The process is as follows:
// 1. We check if there is a JWT token cookie. If there isn't we don't do
//    anything, if there is we go to the next step.
// 2. We validate the token by calling the API /user/validate endpoint.
// 3. If there are any errors we don't change anything and just call the
//    next handler.
// 4. If the API says that the token is not valid we clear the session and
//    remove the token cookie.
// 5. If the token is valid we update the session with the username and call
//    the next handler.
func (h *WebServer) handleSession(c *gin.Context) {
	// We call the next handler no matter what, even if we encounter some
	// errors here.
	defer c.Next()

	session := sessions.Default(c)
	// Check if there is a JWT token cookie, if there isn't, don't do anything
	token, _ := c.Cookie("token")
	if token != "" {
		// Validate the token by calling the API /user/validate endpoint
		resp, err := http.Post(h.Options.ApiURL+"/user/validate", "text/plain", bytes.NewBuffer([]byte(token)))
		// In case of any errors we don't change anything and just call the
		// next handler
		if err != nil {
			log.Println("handleSession: error talking to API: ", err)
			return
		}
		// If the API says that the token is not valid we clear the session and
		// remove the token cookie so that we don't have to all of this again
		if resp.StatusCode != http.StatusOK {
			session.Clear()
			session.Save()
			c.SetCookie("token", "", -1, "/", "localhost", false, true)
			return
		}
		var data auth.UserInfo
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("handleSession: failed to read API response body: ", err)
			return
		}
		if err := json.Unmarshal(b, &data); err != nil {
			log.Println("handleSession: failed to parse API response", err)
			return
		}
		// If the token is valid we update the session with the username
		session.Set("username", data.Username)
		session.Save()
		c.Set("username", data.Username)
	}
}

// handleRoot returns the home page and should be bound to the '/' URL.
// It assumes that there is a template named "index.html" and that it
// was already loaded.
func (h *WebServer) handleRoot(c *gin.Context) {
	username, _ := c.Get("username")

	c.HTML(
		http.StatusOK,
		"index.html",
		gin.H{
			"title":    "Go PB - Home",
			"username": username,
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
	username, _ := c.Get("username")
	c.HTML(
		http.StatusOK,
		"login.html",
		gin.H{
			"title":    "Go PB - Login",
			"errorMsg": "",
			"username": username,
		},
	)
}

// handleDoUserLogin recieves login form data and calls the user API to
// authenticate the user. If successful, it sets the token cookie and
// redirects to the home page.
func (h *WebServer) handleDoUserLogin(c *gin.Context) {
	var u auth.UserLogin
	// Try to parse the form
	if err := c.ShouldBind(&u); err != nil {
		log.Println("handleDoUserLogin: failed to bind to form data: ", err)
		c.Set("errorCode", http.StatusBadRequest)
		c.Set("errorText", http.StatusText(http.StatusBadRequest))
		h.showError(c)
		return
	}
	// Call the API to login
	user, _ := json.Marshal(u)
	resp, err := http.Post(h.Options.ApiURL+"/user/login", "application/json", bytes.NewBuffer(user))

	if err != nil {
		log.Println("handleDoUserLogin: error talking to API: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	// Check if API responded with NotAuthorized
	if resp.StatusCode == http.StatusUnauthorized {
		log.Println("handleDoUserLogin: API returned: ", resp.StatusCode)
		c.HTML(
			resp.StatusCode,
			"login.html",
			gin.H{
				"title":    "Go PB - Login",
				"errorMsg": "Either username or password is incorrect",
			},
		)
		return
	}
	// Check if API responded with some other error
	if resp.StatusCode != http.StatusOK {
		log.Println("handleDoUserLogin: API returned an error: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	// Get API response body and try to parse it as JSON
	var data auth.UserInfo
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("handleDoUserLogin: failed to read API response body: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	if err := json.Unmarshal(b, &data); err != nil {
		log.Println("handleDoUserLogin: failed to parse API response", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	c.SetCookie("token", data.Token, 24*3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}

// handleDoUserLogout logs the user out by clearing the session and the
// token cookie. It redirects to the home page after that.
func (h *WebServer) handleDoUserLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}

// handleUserRegister returns a page with the registration form. It assumes
// that there is a template named "register.html" and that it was already
// loaded.
func (h *WebServer) handleUserRegister(c *gin.Context) {
	username, _ := c.Get("username")
	c.HTML(
		http.StatusOK,
		"register.html",
		gin.H{
			"title":    "Go PB - Register",
			"errorMsg": "",
			"username": username,
		},
	)
}

// handleDoUserRegister recieves the registration form data and calls the user
// API to create new user. If successful it redirects to the login page.
func (h *WebServer) handleDoUserRegister(c *gin.Context) {
	var u auth.UserRegister
	// Try to parse the form
	if err := c.ShouldBind(&u); err != nil {
		log.Println("handleDoUserRegister: failed to bind to form data: ", err)
		c.Set("errorCode", http.StatusBadRequest)
		c.Set("errorText", http.StatusText(http.StatusBadRequest))
		h.showError(c)
		return
	}
	// Call the API to login
	user, _ := json.Marshal(u)
	resp, err := http.Post(h.Options.ApiURL+"/user/register", "application/json", bytes.NewBuffer(user))

	if err != nil {
		log.Println("handleDoUserRegister: error talking to API: ", err)
		c.Set("errorCode", http.StatusInternalServerError)
		c.Set("errorText", http.StatusText(http.StatusInternalServerError))
		c.Set("errorMessage", "Oops! It looks like something went wrong. Don't worry, we have notified the authorities.")
		h.showError(c)
		return
	}
	// Check API response status
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusConflict {
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("handleDoUserRegister: failed to read API response body: ", err)
				b = []byte("")
			}
			msg := []rune(string(b))
			msg[0] = unicode.ToUpper(msg[0])
			c.HTML(
				http.StatusConflict,
				"register.html",
				gin.H{
					"title":    "Go PB - Register",
					"errorMsg": string(msg),
				},
			)
			return
		}
		log.Println("handleDoUserRegister: API returned: ", resp.StatusCode)
		c.Set("errorCode", resp.StatusCode)
		c.Set("errorText", http.StatusText(resp.StatusCode))
		h.showError(c)
		return
	}
	c.Redirect(http.StatusFound, "/u/login")
}

// handlePaste queries the API for a paste and returns a page that displays it.
// It uses the "view.html" template.
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
	username, _ := c.Get("username")
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"Title":    p.Title,
			"Body":     p.Body,
			"Language": p.Syntax,
			"URL":      p.URL(),
			"Server":   "http://localhost:8080", //TODO: this has to come from somewhere
			"username": username,
		},
	)
}

// handlePasteCreate collects information from the new paste form and calls
// the API to create a new paste. If successful it shows the new paste.
// It uses the "view.html" template.
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
	resp, err := http.Post(h.Options.ApiURL+"/paste", "application/json", bytes.NewBuffer(paste))

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
	username, _ := c.Get("username")
	c.HTML(
		http.StatusOK,
		"view.html",
		gin.H{
			"Title":    data.Title,
			"Body":     data.Body,
			"Language": data.Syntax,
			"URL":      resp.Header.Get("Location"),
			"Server":   "http://localhost:8080", //TODO: this has to come from somewhere
			"username": username,
		},
	)
}

// showError displays a custom error page using error.html template.
// The context can use "errorCode", "errorText" and "errorMessage" keys to
// customise what is shown on the page.
// It uses the "error.html" template.
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

	username, _ := c.Get("username")
	c.HTML(
		errorCode,
		"error.html",
		gin.H{
			"title":        "Error",
			"errorCode":    errorCode,
			"errorText":    errorText,
			"errorMessage": errorMsg,
			"username":     username,
		},
	)
}
