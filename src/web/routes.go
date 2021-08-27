// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

package web

import (
	"bytes"
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/go-pkgz/auth/token"
	"github.com/gorilla/mux"
	"github.com/iliafrenkel/go-pb/src/service"
	"github.com/iliafrenkel/go-pb/src/store"
)

// Paginator struct used to build paginators on list pages.
type Paginator struct {
	Number    int
	Offset    int
	Size      int
	IsCurrent bool
}

// PageData contains the data that any page template may need.
type PageData struct {
	Title        string
	Brand        string
	Tagline      string
	Logo         string
	Theme        string
	ID           string
	User         token.User
	Pastes       []store.Paste
	Paste        store.Paste
	Pages        []Paginator
	Server       string
	Version      string
	ErrorCode    int
	ErrorText    string
	ErrorMessage string
	PastesCount  int64
	UsersCount   int64
}

// Generate HTML from a template with PageData.
func (h *Server) generateHTML(tpl string, p PageData) []byte {
	var html bytes.Buffer
	pcnt, ucnt := h.service.GetTotals()
	var pd = PageData{
		Title:        h.options.BrandName + " - " + p.Title,
		Brand:        h.options.BrandName,
		Tagline:      h.options.BrandTagline,
		Logo:         h.options.Logo,
		Theme:        h.options.BootstrapTheme,
		ID:           p.ID,
		User:         p.User,
		Pastes:       p.Pastes,
		Paste:        p.Paste,
		Pages:        p.Pages,
		Server:       h.options.Proto + "://" + h.options.Addr,
		Version:      h.options.Version,
		ErrorCode:    p.ErrorCode,
		ErrorText:    p.ErrorText,
		ErrorMessage: p.ErrorMessage,
		PastesCount:  pcnt,
		UsersCount:   ucnt,
	}

	err := h.templates.ExecuteTemplate(&html, tpl, pd)
	if err != nil {
		h.log.Logf("ERROR error executing template: %v", err)
	}

	return html.Bytes()
}

func (h *Server) showInternalError(w http.ResponseWriter, err error) {
	h.log.Logf("ERROR : %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	_, e := w.Write(h.generateHTML("error.html", PageData{
		Title:        "Error",
		ErrorCode:    http.StatusInternalServerError,
		ErrorText:    http.StatusText(http.StatusInternalServerError),
		ErrorMessage: "",
	}))
	if err != nil {
		h.log.Logf("ERROR showInternalError: failed to write: %v", e)
	}
}

// handleGetHomePage shows the homepage in response to a GET / request.
func (h *Server) handleGetHomePage(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)

	pastes, err := h.service.UserPastes(usr.ID)
	if err != nil {
		h.showInternalError(w, err)
		return
	}

	_, e := w.Write(h.generateHTML("index.html", PageData{Title: "Home", Pastes: pastes, User: usr}))
	if err != nil {
		h.log.Logf("ERROR handleGetHomePage: failed to write: %v", e)
	}
}

// handlePostPaste creates new paste from the form data
func (h *Server) handlePostPaste(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)
	// Read the form data
	r.Body = http.MaxBytesReader(w, r.Body, h.options.MaxBodySize)
	if err := r.ParseForm(); err != nil {
		h.log.Logf("WARN parsing form failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		_, e := w.Write(h.generateHTML("error.html", PageData{
			Title:        "Error",
			ErrorCode:    http.StatusBadRequest,
			ErrorText:    http.StatusText(http.StatusBadRequest),
			ErrorMessage: "",
		}))
		if e != nil {
			h.log.Logf("ERROR handlePostPaste: failed to write: %v", e)
		}
		return
	}
	// Update the user
	_, err := h.service.GetOrUpdateUser(store.User{
		ID:    usr.ID,
		Name:  usr.Name,
		Email: usr.Email,
		IP:    usr.IP,
		Admin: usr.IsAdmin(),
	})
	if err != nil {
		h.log.Logf("ERROR can't update the user: %v", err)
	}
	// Create a new paste
	var p = service.PasteRequest{
		Title:           r.PostFormValue("title"),
		Body:            r.PostFormValue("body"),
		Expires:         r.PostFormValue("expires"),
		DeleteAfterRead: r.PostFormValue("delete_after_read") == "yes",
		Privacy:         r.PostFormValue("privacy"),
		Password:        r.PostFormValue("password"),
		Syntax:          r.PostFormValue("syntax"),
		UserID:          usr.ID,
	}
	paste, err := h.service.NewPaste(p)
	if err != nil {
		if errors.Is(err, service.ErrEmptyBody) {
			w.WriteHeader(http.StatusBadRequest)
			_, e := w.Write(h.generateHTML("error.html", PageData{
				Title:        "Error",
				ErrorCode:    http.StatusBadRequest,
				ErrorText:    http.StatusText(http.StatusBadRequest),
				ErrorMessage: "Body must not be empty.",
			}))
			if e != nil {
				h.log.Logf("ERROR handlePostPaste: failed to write: %v", e)
			}
			return
		}
		if errors.Is(err, service.ErrWrongPrivacy) {
			w.WriteHeader(http.StatusBadRequest)
			_, e := w.Write(h.generateHTML("error.html", PageData{
				Title:        "Error",
				ErrorCode:    http.StatusBadRequest,
				ErrorText:    http.StatusText(http.StatusBadRequest),
				ErrorMessage: "Privacy can be one of 'private', 'public' or 'unlisted'.",
			}))
			if e != nil {
				h.log.Logf("ERROR handlePostPaste: failed to write: %v", e)
			}
			return
		}
		if errors.Is(err, service.ErrWrongDuration) {
			w.WriteHeader(http.StatusBadRequest)
			_, e := w.Write(h.generateHTML("error.html", PageData{
				Title:        "Error",
				ErrorCode:    http.StatusBadRequest,
				ErrorText:    http.StatusText(http.StatusBadRequest),
				ErrorMessage: "Duration format is incorrect.",
			}))
			if e != nil {
				h.log.Logf("ERROR handlePostPaste: failed to write: %v", e)
			}
			return
		}
		// Some bad thing happened and we don't know what to do
		h.showInternalError(w, err)
		return
	}
	// Get a list of user pastes
	pastes, err := h.service.UserPastes(usr.ID)
	if err != nil {
		h.showInternalError(w, err)
		return
	}

	_, e := w.Write(h.generateHTML("view.html", PageData{
		Title:  "Paste",
		Pastes: pastes,
		Paste:  paste,
		User:   usr,
	}))
	if e != nil {
		h.log.Logf("ERROR handlePostPaste: failed to write: %v", e)
	}
}

// handleGetPastePage generates a page to view a single paste.
func (h *Server) handleGetPastePage(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)
	// Get paste encoded ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		h.log.Logf("WARN handleGetPastePage: paste id not found")
		w.WriteHeader(http.StatusBadRequest)
		_, e := w.Write(h.generateHTML("error.html", PageData{
			Title:        "Error",
			ErrorCode:    http.StatusBadRequest,
			ErrorText:    http.StatusText(http.StatusBadRequest),
			ErrorMessage: "",
		}))
		if e != nil {
			h.log.Logf("ERROR handleGetPastePage: failed to write: %v", e)
		}
		return
	}
	// If the request comes from a password form, get the password
	if err := r.ParseForm(); err != nil {
		h.log.Logf("WARN parsing form failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		_, e := w.Write(h.generateHTML("error.html", PageData{
			Title:        "Error",
			ErrorCode:    http.StatusBadRequest,
			ErrorText:    http.StatusText(http.StatusBadRequest),
			ErrorMessage: "",
		}))
		if e != nil {
			h.log.Logf("ERROR handleGetPastePage: failed to write: %v", e)
		}
		return
	}
	pwd := r.PostFormValue("password")

	// Get the paste from the storage
	paste, err := h.service.GetPaste(id, usr.ID, pwd)
	if err != nil {
		// Check if paste was not found
		if errors.Is(err, service.ErrPasteNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_, e := w.Write(h.generateHTML("error.html", PageData{
				Title:        "Error",
				ErrorCode:    http.StatusNotFound,
				ErrorText:    http.StatusText(http.StatusNotFound),
				ErrorMessage: "There is no such paste",
			}))
			if e != nil {
				h.log.Logf("ERROR handleGetPastePage: failed to write: %v", e)
			}
			return
		}
		// Check if paste is private an belongs to another user
		if errors.Is(err, service.ErrPasteIsPrivate) {
			w.WriteHeader(http.StatusForbidden)
			_, e := w.Write(h.generateHTML("error.html", PageData{
				Title:        "Error",
				ErrorCode:    http.StatusForbidden,
				ErrorText:    http.StatusText(http.StatusForbidden),
				ErrorMessage: "This paste is private",
			}))
			if e != nil {
				h.log.Logf("ERROR handleGetPastePage: failed to write: %v", e)
			}
			return
		}
		// Check if paste is password-protected
		if errors.Is(err, service.ErrPasteHasPassword) || errors.Is(err, service.ErrWrongPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			_, e := w.Write(h.generateHTML("password.html", PageData{
				ID:           id,
				User:         usr,
				Title:        "Password",
				ErrorMessage: "This paste is protected by a password",
			}))
			if e != nil {
				h.log.Logf("ERROR handleGetPastePage: failed to write: %v", e)
			}
			return
		}
		// Some other error that we didn't expect
		h.showInternalError(w, err)
		return
	}

	// Get user pastes
	pastes, err := h.service.UserPastes(usr.ID)
	if err != nil {
		h.showInternalError(w, err)
		return
	}

	_, e := w.Write(h.generateHTML("view.html", PageData{
		Title:  "Paste",
		Pastes: pastes,
		Paste:  paste,
		User:   usr,
	}))
	if e != nil {
		h.log.Logf("ERROR handleGetPastePage: failed to write: %v", e)
	}
}

// handleGetPastesList generates a page to view a list of pastes.
func (h *Server) handleGetPastesList(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)
	limit := 10 //TODO: make it configurable as PageSize
	skip, err := strconv.Atoi(r.FormValue("skip"))
	if err != nil {
		skip = 0
	}

	pastes, err := h.service.GetPastes(usr.ID, "-created", limit, skip)
	if err != nil {
		h.showInternalError(w, err)
		return
	}
	count := h.service.PastesCount(usr.ID)
	pageCount := int(math.Ceil(float64(count) / float64(limit)))

	pages := make([]Paginator, pageCount)
	for i := 1; i <= pageCount; i++ {
		pages[i-1] = Paginator{
			Number:    i,
			Offset:    (i - 1) * limit,
			Size:      limit,
			IsCurrent: skip/limit == i-1,
		}
	}

	_, e := w.Write(h.generateHTML("list.html", PageData{
		Title:  "Pastes",
		Pastes: pastes,
		Pages:  pages,
		User:   usr,
	}))
	if e != nil {
		h.log.Logf("ERROR handleGetPastesList: failed to write: %v", e)
	}
}

// Show 404 Not Found error page
func (h *Server) notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, e := w.Write(h.generateHTML("error.html", PageData{
		Title:        "Error",
		ErrorCode:    http.StatusNotFound,
		ErrorText:    http.StatusText(http.StatusNotFound),
		ErrorMessage: "Unfortunately the page you are looking for is not there 🙁",
	}))
	if e != nil {
		h.log.Logf("ERROR notFound: failed to write: %v", e)
	}
}
