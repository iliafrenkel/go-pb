package http

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/iliafrenkel/go-pb/api/api"
)

type ApiHandler struct {
	PasteService api.PasteService
	Router       *mux.Router
}

func New(svc api.PasteService) *ApiHandler {
	var handler ApiHandler

	handler.PasteService = svc
	handler.Router = mux.NewRouter()
	handler.Router.HandleFunc("/paste/{id}", handler.handlePaste).Methods("GET")
	handler.Router.HandleFunc("/create", handler.handleCreate).Methods("POST")
	handler.Router.HandleFunc("/delete/{id}", handler.handleDelete).Methods("PUT")

	return &handler
}

func (h *ApiHandler) handlePaste(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p, err := h.PasteService.Paste(vars["id"])
	if err != nil {
		fmt.Println(fmt.Errorf("paste %s not found", vars["id"]))
		w.WriteHeader(http.StatusNotFound)
		return
	}
	res, err := json.Marshal(p)
	if err != nil {
		fmt.Println(fmt.Errorf("error converting paste to json: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", res)
}

func (h *ApiHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
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
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
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
		fmt.Println(fmt.Errorf("failed to create paste: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(p)
	if err != nil {
		fmt.Println(fmt.Errorf("error converting paste to json: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", res)
}

func (h *ApiHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := h.PasteService.Delete(vars["id"]); err != nil {
		fmt.Println(fmt.Errorf("paste %s not found", vars["id"]))
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (h *ApiHandler) ListenAndServe(addr string) error {
	// TODO: Implement timeouts and graceful shutdown as per
	// https://github.com/gorilla/mux#graceful-shutdown
	return http.ListenAndServe(addr, h.Router)
}
