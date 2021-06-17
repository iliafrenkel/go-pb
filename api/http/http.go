package http

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
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
	// TODO: needs improvement as per
	// https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	var data struct {
		Title string
		Body  []byte
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		fmt.Println(fmt.Errorf("failed to parse the form: %v", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Println(data)

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
