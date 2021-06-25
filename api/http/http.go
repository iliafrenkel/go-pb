package http

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/iliafrenkel/go-pb/api"
)

// ApiServer type provides an HTTP server that calls PasteService methods in
// response to HTTP requests to certain routes.
//
// Use the `New` function to create an instance of ApiServer with the default
// routes.
type ApiServer struct {
	PasteService api.PasteService
	Router       *mux.Router
}

// New function returns an instance of ApiServer using provided PasteService
// and the default HTTP routes for manipulating pastes.
//
// The routes are:
//   GET    /paste/{id} - get paste by ID
//   POST   /paste      - create new paste
//   DELETE /paste/{id} - delete paste by ID
func New(svc api.PasteService) *ApiServer {
	var handler ApiServer

	handler.PasteService = svc
	handler.Router = mux.NewRouter()
	handler.Router.HandleFunc("/paste/{id}", handler.handlePaste).Methods("GET")
	handler.Router.HandleFunc("/paste", handler.handleCreate).Methods("POST")
	handler.Router.HandleFunc("/paste/{id}", handler.handleDelete).Methods("DELETE")

	return &handler
}

// ListenAndServe starts an HTTP server and binds it to the provided address.
//
// The server is configured with timeouts and graceful shutdown as per
// https://github.com/gorilla/mux#graceful-shutdown
//
// TODO: Timeouts should be configurable.
func (h *ApiServer) ListenAndServe(addr string) {
	var wait time.Duration = time.Second * 15 // shutdown timeout

	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h.Router,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
}

// handlePaste is an HTTP handler for the GET /paste/{id} route, it returns
// the paste as a JSON string or 404 Not Found.
func (h *ApiServer) handlePaste(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, err := h.PasteService.Paste(vars["id"])
	if err != nil {
		http.Error(w, "paste not found", http.StatusNotFound)
		return
	}
	res, err := json.Marshal(p)
	if err != nil {
		log.Printf("error converting paste to json: %v\n", err)
		http.Error(w, "error converting paste to json", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", res)
}

// handleCreate is an HTTP handler for the POST /paste route. It expects the
// new paste as a JSON sting in the body of the request. Returns newly created
// paste as a JSON string.
//
// The JSON object must correspond to the api.Paste struct. Absent fields will
// get default values. Extra fields will generate an error. Only one object is
// expected, multiple JSON objects in the body will result in an error. Body
// size is currently limited to a hardcoded value of 10KB.
//
// TODO: Make maximum body size configurable.
func (h *ApiServer) handleCreate(w http.ResponseWriter, r *http.Request) {
	// Parse incoming json
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

	// If the Content-Type header is present, check that it has the value
	// application/json.
	if h := r.Header.Get("Content-Type"); h != "" {
		if h != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 10KB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 10240)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var data api.Paste
	if err := dec.Decode(&data); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains malformed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 10KB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Create new paste
	b := make([]byte, 16)
	rand.Read(b)
	p := api.Paste{
		ID:      fmt.Sprintf("%x", md5.Sum(b)),
		Title:   data.Title,
		Body:    data.Body,
		Expires: time.Time{},
	}
	if err := h.PasteService.Create(&p); err != nil {
		log.Printf("failed to create paste: %v\n", err)
		http.Error(w, "failed to create paste", http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(p)
	if err != nil {
		log.Printf("error converting paste to json: %v\n", err)
		http.Error(w, "error converting paste to json", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", res)
}

// handleDelete is an HTTP handler for the DELETE /paste/{id} route. Returns
// 200 OK or 404 Not Found.
func (h *ApiServer) handleDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.PasteService.Delete(vars["id"]); err != nil {
		http.Error(w, "paste not found", http.StatusNotFound)
		return
	}
}
