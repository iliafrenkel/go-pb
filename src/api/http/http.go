// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.package main
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/base62"
)

// ApiServerOptions defines various parameters needed to run the ApiServer
type ApiServerOptions struct {
	// Addr will be passed to http.Server to listen on, see http.Server
	// documentation for more information.
	Addr string
	// Maximum size of the POST request body, anything larger than this will
	// be rejected with an error.
	MaxBodySize int64
}

// ApiServer type provides an HTTP server that calls PasteService methods in
// response to HTTP requests to certain routes.
//
// Use the `New` function to create an instance of ApiServer with the default
// routes.
type ApiServer struct {
	PasteService api.PasteService
	UserService  api.UserService
	Router       *gin.Engine
	Server       *http.Server
	Options      ApiServerOptions
}

// New function returns an instance of ApiServer using provided PasteService
// and all the HTTP routes for manipulating pastes.
//
// The routes are:
//   GET    /paste/{id}    - get paste by ID
//   POST   /paste         - create new paste
//   DELETE /paste/{id}    - delete paste by ID
//   POST   /user/login    - authenticate user
//   POST   /user/register - register new user
func New(pSvc api.PasteService, uSvc api.UserService, opts ApiServerOptions) *ApiServer {
	var handler ApiServer
	handler.Options = opts

	handler.PasteService = pSvc
	handler.UserService = uSvc

	handler.Router = gin.Default()

	paste := handler.Router.Group("/paste")
	{
		paste.GET("/:id", handler.handlePaste)
		paste.POST("", handler.verifyJsonMiddleware(new(api.PasteForm)), handler.handleCreate)
		paste.DELETE("/:id", handler.handleDelete)
		paste.GET("/list", handler.handleListPaste)
	}

	user := handler.Router.Group("/user")
	{
		user.POST("/login", handler.verifyJsonMiddleware(new(api.UserLogin)), handler.handleUserLogin)
		user.POST("/register", handler.verifyJsonMiddleware(new(api.UserRegister)), handler.handleUserRegister)
		user.POST("/validate", handler.verifyJsonMiddleware(new(api.UserInfo)), handler.handleUserValidate)
	}

	return &handler
}

// ListenAndServe starts an HTTP server and binds it to the provided address.
//
// TODO: Timeouts should be configurable.
func (h *ApiServer) ListenAndServe() error {
	h.Server = &http.Server{
		Addr: h.Options.Addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
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

func (h *ApiServer) verifyJsonMiddleware(data interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse incoming json
		// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

		// If the Content-Type header is present, check that it has the value
		// application/json.
		if hdr := c.GetHeader("Content-Type"); hdr != "" {
			if hdr != "application/json" {
				c.String(http.StatusUnsupportedMediaType, "wrong Content-Type header, expect application/json")
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
				c.String(http.StatusBadRequest, fmt.Sprintf("request body contains malformed JSON (at position %d)", syntaxError.Offset))

			// In some circumstances Decode() may also return an
			// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
			// is an open issue regarding this at
			// https://github.com/golang/go/issues/25956.
			case errors.Is(err, io.ErrUnexpectedEOF):
				c.String(http.StatusBadRequest, "request body contains malformed JSON")

			// Catch any type errors, like trying to assign a string in the
			// JSON request body to a int field in our Paste struct. We can
			// interpolate the relevant field name and position into the error
			// message to make it easier for the client to fix.
			case errors.As(err, &unmarshalTypeError):
				c.String(http.StatusBadRequest, fmt.Sprintf("request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset))

			// Catch the error caused by extra unexpected fields in the request
			// body. We extract the field name from the error message and
			// interpolate it in our custom error message. There is an open
			// issue at https://github.com/golang/go/issues/29035 regarding
			// turning this into a sentinel error.
			case strings.HasPrefix(err.Error(), "json: unknown field "):
				fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
				c.String(http.StatusBadRequest, fmt.Sprintf("request body contains unknown field %s", fieldName))

			// An io.EOF error is returned by Decode() if the request body is
			// empty.
			case errors.Is(err, io.EOF):
				c.String(http.StatusBadRequest, "request body must not be empty")

			// Catch the error caused by the request body being too large. Again
			// there is an open issue regarding turning this into a sentinel
			// error at https://github.com/golang/go/issues/30715.
			case err.Error() == "http: request body too large":
				c.String(http.StatusBadRequest, fmt.Sprintf("request body must not be larger than %d bytes", h.Options.MaxBodySize))

			// Otherwise default to logging the error and sending a 500 Internal
			// Server Error response.
			default:
				log.Println("verifyJsonMiddleware: ", err.Error())
				c.String(http.StatusInternalServerError, err.Error())
			}
			c.Abort()
			return
		}

		// Call decode again, using a pointer to an empty anonymous struct as
		// the destination. If the request body only contained a single JSON
		// object this will return an io.EOF error. So if we get anything else,
		// we know that there is additional data in the request body.
		if err := dec.Decode(&struct{}{}); err != io.EOF {
			c.String(http.StatusInternalServerError, "request body must only contain a single JSON object")
			c.Abort()
			return
		}

		c.Set("payload", data)
		c.Next()
	}
}

// handlePaste is an HTTP handler for the GET /paste/{id} route, it returns
// the paste as a JSON string or 404 Not Found.
func (h *ApiServer) handlePaste(c *gin.Context) {
	// We expect the id parameter as base62 encoded string, we try to decode
	// it into a uint64 paste id and return 404 if we can't.
	id, err := base62.Decode(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.String(http.StatusNotFound, "paste not found")
		return
	}

	p, err := h.PasteService.Get(id)
	if err != nil {
		log.Println(err)
		c.String(http.StatusNotFound, "paste not found")
		return
	}

	c.JSON(http.StatusOK, p)

	// We "burn" the paste if DeleteAfterRead flag is set.
	if p.DeleteAfterRead {
		h.PasteService.Delete(p.ID)
	}
}

// handleCreate is an HTTP handler for the POST /paste route. It expects the
// new paste as a JSON sting in the body of the request. Returns newly created
// paste as a JSON string and the 'Location' header set to the new paste URL.
//
// The JSON object must correspond to the api.Paste struct. Absent fields will
// get default values. Extra fields will generate an error. Only one object is
// expected, multiple JSON objects in the body will result in an error. Body
// size is currently limited to a configurable value of Options.MaxBodySize.
func (h *ApiServer) handleCreate(c *gin.Context) {
	data := c.MustGet("payload").(*api.PasteForm)

	p, err := h.PasteService.Create(*data)
	if err != nil {
		log.Printf("handleCreate: failed to create paste: %v\n", err)
		c.String(http.StatusBadRequest, "failed to create paste")
		return
	}
	c.Header("Location", p.URL())
	c.JSON(http.StatusCreated, p)
}

// handleDelete is an HTTP handler for the DELETE /paste/{id} route. Deletes
// the paste by id and returns 200 OK or 404 Not Found.
func (h *ApiServer) handleDelete(c *gin.Context) {
	id, err := base62.Decode(c.Param("id"))
	if err != nil {
		c.String(http.StatusNotFound, "paste not found")
		return
	}

	if err := h.PasteService.Delete(id); err != nil {
		c.String(http.StatusNotFound, "paste not found")
		return
	}
}

// handleListPaste is an HTTP handlers for GET /paste/list route. Returns
// an array of all pastes.
func (h *ApiServer) handleListPaste(c *gin.Context) {
	pastes := h.PasteService.List()

	c.JSON(http.StatusOK, pastes)
}

// handleUserLogin is an HTTP handler for POST /user/login route. It returns
// auth.UserInfo with the username and JWT token on success.
func (h *ApiServer) handleUserLogin(c *gin.Context) {
	data := c.MustGet("payload").(*api.UserLogin)

	// Login returns Username and JWT token
	var usr api.UserInfo
	usr, err := h.UserService.Authenticate(*data)
	if err != nil {
		log.Printf("failed to login: %v\n", err)
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	c.JSON(http.StatusOK, usr)
}

// handleUserRegister is an HTTP handler for POST /user/register route. It
// tries to create a new user and returns 200 OK on success.
func (h *ApiServer) handleUserRegister(c *gin.Context) {
	data := c.MustGet("payload").(*api.UserRegister)

	// Register doesn't return anything
	err := h.UserService.Create(*data)
	if err != nil {
		log.Printf("failed to create new user: %v\n", err)
		var msg = struct{ Message string }{Message: err.Error()}
		c.JSON(http.StatusConflict, msg)
		return
	}

	c.Status(http.StatusOK)
}

// handleUserValidate verifyes that the JWT token is correct
func (h *ApiServer) handleUserValidate(c *gin.Context) {
	data := c.MustGet("payload").(*api.UserInfo)

	usr, err := h.UserService.Validate(api.User{}, data.Token)
	if err != nil {
		log.Printf("handleUserValidate: validation failed: %v", err.Error())
		c.Status(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, usr)
}
