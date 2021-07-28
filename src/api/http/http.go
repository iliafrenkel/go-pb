// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package http provides an ApiServer type - a server that uses
// api.PasteService and api.UserService to provide many useful endpoints.
// Check the New method documentation for the list of all endpoints.
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/base62"
	"golang.org/x/crypto/bcrypt"
)

// APIServerOptions defines various parameters needed to run the ApiServer
type APIServerOptions struct {
	// Addr will be passed to http.Server to listen on, see http.Server
	// documentation for more information.
	Addr string
	// Maximum size of the POST request body, anything larger than this will
	// be rejected with an error.
	MaxBodySize int64
	// When using a database as a storage this connection string will be passed
	// on to the corresponding service.
	DBConnectionString string
	//Read timeout: maximum duration for reading the entire request.
	ReadTimeout time.Duration
	// Write timeout: maximum duration before timing out writes of the response
	WriteTimeout time.Duration
	// Idle timeout: maximum amount of time to wait for the next request
	IdleTimeout time.Duration
	// Log file location
	LogFile string
	// Log mode 'debug' or 'production'
	LogMode string
	//
	DBAutoMigrate bool
	//
	TokenSecret string
}

// APIServer type provides an HTTP server that calls PasteService methods in
// response to HTTP requests to certain routes.
//
// Use the `New` function to create an instance of ApiServer with the default
// routes.
type APIServer struct {
	PasteService api.PasteService
	UserService  api.UserService
	Router       *gin.Engine
	Server       *http.Server
	Options      APIServerOptions
}

// New function returns an instance of ApiServer using provided PasteService
// and all the HTTP routes for manipulating pastes.
//
// The routes are:
//   GET    /paste/{id}      - get paste by ID
//   POST   /paste/{id}      - get password protected paste by ID
//   POST   /paste           - create new paste
//   DELETE /paste/{id}      - delete paste by ID
//   GET    /paste/list/{id} - get a list of pastes by UserID
//   POST   /user/login      - authenticate user
//   POST   /user/register   - register new user
//   POST   /user/validate   - validate user token
func New(pSvc api.PasteService, uSvc api.UserService, opts APIServerOptions) *APIServer {
	var handler APIServer
	handler.Options = opts

	handler.PasteService = pSvc
	handler.UserService = uSvc

	if handler.Options.LogMode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if handler.Options.LogFile != "" {
		gin.DisableConsoleColor()
		f, err := os.OpenFile(handler.Options.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	}

	handler.Router = gin.New()
	handler.Router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("\033[97;41m[API]\033[0m %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCodeColor(), param.StatusCode, param.ResetColor(),
			param.Latency,
			param.ClientIP,
			param.MethodColor(), param.Method, param.ResetColor(),
			param.Path,
			param.ErrorMessage,
		)
	}), gin.Recovery())

	paste := handler.Router.Group("/paste")
	{
		paste.GET("/:id", handler.handlePasteGet)
		paste.POST("/:id", handler.verifyJSONMiddleware(new(api.PastePassword)), handler.handlePasteGetWithPassword)
		paste.POST("", handler.verifyJSONMiddleware(new(api.PasteForm)), handler.handlePasteCreate)
		paste.DELETE("/:id", handler.handlePasteDelete)
		paste.GET("/list/:id", handler.handlePasteList)
	}

	user := handler.Router.Group("/user")
	{
		user.POST("/login", handler.verifyJSONMiddleware(new(api.UserLogin)), handler.handleUserLogin)
		user.POST("/register", handler.verifyJSONMiddleware(new(api.UserRegister)), handler.handleUserRegister)
		user.POST("/validate", handler.verifyJSONMiddleware(new(api.UserInfo)), handler.handleUserValidate)
	}

	return &handler
}

// ListenAndServe starts an HTTP server and binds it to the provided address.
//
// TODO: Timeouts should be configurable.
func (h *APIServer) ListenAndServe() error {
	// Good practice to set timeouts to avoid Slowloris attacks.
	h.Server = &http.Server{
		Addr:         h.Options.Addr,
		WriteTimeout: h.Options.WriteTimeout,
		ReadTimeout:  h.Options.ReadTimeout,
		IdleTimeout:  h.Options.IdleTimeout,
		Handler:      h.Router,
	}

	return h.Server.ListenAndServe()
}

// verifyPayload checks that the incoming Json payload arrived with the correct
// content type and can be properly decoded.
// The JSON object must correspond to the struct referenced by the "payload"
// context. Absent fields will get default values. Extra fields will generate
// an error. Only one object is expected, multiple JSON objects in the body
// will result in an error. Body size is limited to the value of
// Options.MaxBodySize parameter.
func (h *APIServer) verifyJSONMiddleware(data interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse incoming json
		// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

		// If the Content-Type header is present, check that it has the value
		// application/json.
		if hdr := c.GetHeader("Content-Type"); hdr != "" {
			if hdr != "application/json" {
				c.JSON(http.StatusUnsupportedMediaType, api.HTTPError{
					Code:    http.StatusUnsupportedMediaType,
					Message: fmt.Sprintf("Incorrect Content-Type header [%s], expect [application/json]", hdr),
				})
				c.Abort()
				return
			}
		}

		// Use http.MaxBytesReader to enforce a maximum read size from the
		// response body. A request body larger than that will now result in
		// Decode() returning a "http: request body too large" error.
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.Options.MaxBodySize)

		// Setup the decoder and call the DisallowUnknownFields() method on it.
		// This will cause Decode() to return a "json: unknown field ..." error
		// if it encounters any extra unexpected fields in the JSON. Strictly
		// speaking, it returns an error for "keys which do not match any
		// non-ignored, exported fields in the destination".
		dec := json.NewDecoder(c.Request.Body)
		dec.DisallowUnknownFields()

		if err := dec.Decode(&data); err != nil {
			var syntaxError *json.SyntaxError
			var unmarshalTypeError *json.UnmarshalTypeError
			switch {
			// Catch any syntax errors in the JSON and send an error message
			// which interpolates the location of the problem to make it
			// easier for the client to fix.
			case errors.As(err, &syntaxError):
				c.JSON(http.StatusBadRequest, api.HTTPError{
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("Request body contains malformed JSON (at position %d)", syntaxError.Offset),
				})

			// In some circumstances Decode() may also return an
			// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
			// is an open issue regarding this at
			// https://github.com/golang/go/issues/25956.
			case errors.Is(err, io.ErrUnexpectedEOF):
				c.JSON(http.StatusBadRequest, api.HTTPError{
					Code:    http.StatusBadRequest,
					Message: "Request body contains malformed JSON",
				})

			// Catch any type errors, like trying to assign a string in the
			// JSON request body to a int field in our Paste struct. We can
			// interpolate the relevant field name and position into the error
			// message to make it easier for the client to fix.
			case errors.As(err, &unmarshalTypeError):
				c.JSON(http.StatusBadRequest, api.HTTPError{
					Code: http.StatusBadRequest,
					Message: fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)",
						unmarshalTypeError.Field,
						unmarshalTypeError.Offset),
				})

			// Catch the error caused by extra unexpected fields in the request
			// body. We extract the field name from the error message and
			// interpolate it in our custom error message. There is an open
			// issue at https://github.com/golang/go/issues/29035 regarding
			// turning this into a sentinel error.
			case strings.HasPrefix(err.Error(), "json: unknown field "):
				fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
				c.JSON(http.StatusBadRequest, api.HTTPError{
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("Request body contains unknown field %s", fieldName),
				})

			// An io.EOF error is returned by Decode() if the request body is
			// empty.
			case errors.Is(err, io.EOF):
				c.JSON(http.StatusBadRequest, api.HTTPError{
					Code:    http.StatusBadRequest,
					Message: "Request body must not be empty",
				})

			// Catch the error caused by the request body being too large. Again
			// there is an open issue regarding turning this into a sentinel
			// error at https://github.com/golang/go/issues/30715.
			case err.Error() == "http: request body too large":
				c.JSON(http.StatusBadRequest, api.HTTPError{
					Code:    http.StatusBadRequest,
					Message: fmt.Sprintf("Request body must not be larger than %d bytes", h.Options.MaxBodySize),
				})

			// Otherwise default to logging the error and sending a 500 Internal
			// Server Error response.
			default:
				log.Println("verifyJsonMiddleware: unexpected error: ", err.Error())
				c.JSON(http.StatusInternalServerError, api.HTTPError{
					Code:    http.StatusInternalServerError,
					Message: fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()),
				})
			}
			c.Abort()
			return
		}

		// Call decode again, using a pointer to an empty anonymous struct as
		// the destination. If the request body only contained a single JSON
		// object this will return an io.EOF error. So if we get anything else,
		// we know that there is additional data in the request body.
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			c.JSON(http.StatusBadRequest, api.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "Request body must only contain a single JSON object",
			})
			c.Abort()
			return
		}

		c.Set("payload", data)
		c.Next()
	}
}

// handlePasteGet is an HTTP handler for the GET /paste/{id} route, it returns
// the paste as a JSON string or 404 Not Found. If paste is password protected
// we return 401 Unauthorised and the caller has to POST to /paste/{id} to with
// the password to get the paste.
func (h *APIServer) handlePasteGet(c *gin.Context) {
	// We expect the id parameter as base62 encoded string, we try to decode
	// it into a int64 paste id and return 404 if we can't.
	id, err := base62.Decode(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, api.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Paste not found",
		})
		return
	}

	p, err := h.PasteService.Get(id)
	// Service returned an error and we don't know what to do here. We log the
	// error and send 500 InternalServerError back to the caller.
	if err != nil {
		log.Println("handlePasteGet: unexpected error: ", err.Error())
		c.JSON(http.StatusInternalServerError, api.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()),
		})
		return
	}
	// Service call was successful but returned nil. Respond with 404 NoFound.
	if p == nil {
		c.JSON(http.StatusNotFound, api.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Paste not found",
		})
		return
	}
	// Check if the paste is password protected. If yes, we return 401
	// Unauthorized and the caller must POST back with the password.
	if p.Password != "" {
		c.JSON(http.StatusUnauthorized, api.HTTPError{
			Code:    http.StatusUnauthorized,
			Message: "Paste is password protected",
		})
		return
	}

	// All is good, send the paste back to the caller.
	c.JSON(http.StatusOK, p)

	// We "burn" the paste if DeleteAfterRead flag is set.
	if p.DeleteAfterRead {
		h.PasteService.Delete(p.ID)
	}
}

func (h *APIServer) handlePasteGetWithPassword(c *gin.Context) {
	// We expect the id parameter as base62 encoded string, we try to decode
	// it into a int64 paste id and return 404 if we can't.
	id, err := base62.Decode(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, api.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Paste not found",
		})
		return
	}
	// Get the password from the context, the verifyJSONMiddlerware should've
	// prepared it for us.
	var pwd string
	if data, ok := c.Get("payload"); !ok {
		log.Println("handlePasteGetWithPassword: unexpected error: ", err.Error())
		c.JSON(http.StatusInternalServerError, api.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()),
		})
		return
	} else {
		pwd = data.(*api.PastePassword).Password
	}
	// Get the paste
	p, err := h.PasteService.Get(id)
	// Service returned an error and we don't know what to do here. We log the
	// error and send 500 InternalServerError back to the caller.
	if err != nil {
		log.Println("handlePasteGet: unexpected error: ", err.Error())
		c.JSON(http.StatusInternalServerError, api.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: fmt.Sprintf("%s: %s", http.StatusText(http.StatusInternalServerError), err.Error()),
		})
		return
	}
	// Service call was successful but returned nil. Respond with 404 NoFound.
	if p == nil {
		c.JSON(http.StatusNotFound, api.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Paste not found",
		})
		return
	}
	// Verify the password
	if err := bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(pwd)); err != nil {
		c.JSON(http.StatusUnauthorized, api.HTTPError{
			Code:    http.StatusUnauthorized,
			Message: "Paste password is incorrect",
		})
		return
	}

	// All is good, send the paste back to the caller.
	c.JSON(http.StatusOK, p)

	// We "burn" the paste if DeleteAfterRead flag is set.
	if p.DeleteAfterRead {
		h.PasteService.Delete(p.ID)
	}

}

// handlePasteCreate is an HTTP handler for the POST /paste route. It expects
// the new paste as a JSON sting in the body of the request. Returns newly
// created paste as a JSON string and the 'Location' header set to the new
// paste URL.
//
// The JSON object must correspond to the api.PasteForm struct. Absent fields
// will get default values. Extra fields will generate an error. Only one
// object is expected, multiple JSON objects in the body will result in an
// error. Body size is currently limited to a configurable value of
// Options.MaxBodySize.
func (h *APIServer) handlePasteCreate(c *gin.Context) {
	data := c.MustGet("payload").(*api.PasteForm)

	p, err := h.PasteService.Create(*data)
	if err != nil {
		log.Printf("handleCreate: failed to create paste: %v\n", err)
		c.JSON(http.StatusBadRequest, api.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "Failed to create paste",
		})
		return
	}
	c.Header("Location", p.URL())
	c.JSON(http.StatusCreated, p)
}

// handlePasteDelete is an HTTP handler for the DELETE /paste/:id route. Deletes
// the paste by id and returns 200 OK or 404 Not Found.
func (h *APIServer) handlePasteDelete(c *gin.Context) {
	id, err := base62.Decode(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, api.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Paste not found",
		})
		return
	}

	if err := h.PasteService.Delete(int64(id)); err != nil {
		c.JSON(http.StatusNotFound, api.HTTPError{
			Code:    http.StatusNotFound,
			Message: "Paste not found",
		})
		return
	}
	c.Header("Content-Type", "application/json")
}

// handlePasteList is an HTTP handlers for GET /paste/list/:id route. Returns
// an array of pastes by user ID.
func (h *APIServer) handlePasteList(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "ID is incorrect",
		})
		return
	}
	pastes := h.PasteService.List(id)

	c.JSON(http.StatusOK, pastes)
}

// handleUserLogin is an HTTP handler for POST /user/login route. It returns
// auth.UserInfo with the username and JWT token on success.
func (h *APIServer) handleUserLogin(c *gin.Context) {
	data := c.MustGet("payload").(*api.UserLogin)

	// Login returns Username and JWT token
	var usr api.UserInfo
	usr, err := h.UserService.Authenticate(*data)
	if err != nil {
		log.Printf("failed to login: %v\n", err)
		c.JSON(http.StatusUnauthorized, api.HTTPError{
			Code:    http.StatusUnauthorized,
			Message: "Invalid credentials",
		})
		return
	}

	c.JSON(http.StatusOK, usr)
}

// handleUserRegister is an HTTP handler for POST /user/register route. It
// tries to create a new user and returns 200 OK on success.
func (h *APIServer) handleUserRegister(c *gin.Context) {
	data := c.MustGet("payload").(*api.UserRegister)

	// Register doesn't return anything
	err := h.UserService.Create(*data)
	if err != nil {
		log.Printf("failed to create new user: %v\n", err)
		c.JSON(http.StatusConflict, api.HTTPError{
			Code:    http.StatusConflict,
			Message: err.Error(),
		})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)
}

// handleUserValidate verifyes that the JWT token is correct
func (h *APIServer) handleUserValidate(c *gin.Context) {
	data := c.MustGet("payload").(*api.UserInfo)

	usr, err := h.UserService.Validate(api.User{}, data.Token)
	if err != nil {
		log.Printf("handleUserValidate: validation failed: %v\n", err.Error())
		c.JSON(http.StatusUnauthorized, api.HTTPError{
			Code:    http.StatusUnauthorized,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, usr)
}
