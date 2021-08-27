// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

package web

import (
	"errors"
	"math"
	"net/http"
	"strconv"

	"github.com/go-pkgz/auth/token"
	"github.com/gorilla/mux"
	"github.com/iliafrenkel/go-pb/src/service"
	"github.com/iliafrenkel/go-pb/src/store"
	"github.com/iliafrenkel/go-pb/src/web/page"
)

// showInternalError writes 500 Internal Server Error page.
func (h *Server) showInternalError(w http.ResponseWriter, err error) {
	h.log.Logf("ERROR : %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	p := page.New(h.templates,
		page.Template("error.html"),
		page.Title(h.options.BrandName+" - Error"),
		page.ErrorCode(http.StatusInternalServerError),
		page.ErrorText(http.StatusText(http.StatusInternalServerError)),
	)

	e := p.Show(w)
	if e != nil {
		h.log.Logf("ERROR showInternalError: failed to generate page: %v", e)
	}
}

// showError writes an error page.
func (h *Server) showError(w http.ResponseWriter, httpError int, msg string) {
	w.WriteHeader(httpError)
	p := page.New(h.templates,
		page.Template("error.html"),
		page.Title(h.options.BrandName+" - Error"),
		page.ErrorCode(httpError),
		page.ErrorText(http.StatusText(httpError)),
		page.ErrorMessage(msg),
	)

	e := p.Show(w)
	if e != nil {
		h.log.Logf("ERROR showError: failed to generate page: %v", e)
	}
}

// showPage generates a page and writes to the response
func (h *Server) showPage(w http.ResponseWriter, data ...page.Data) {
	pastes, users := h.service.GetTotals()
	totals := page.Stats{
		Pastes: pastes,
		Users:  users,
	}
	p := page.New(h.templates,
		page.Brand(h.options.BrandName),
		page.Tagline(h.options.BrandTagline),
		page.Logo(h.options.Logo),
		page.Theme(h.options.BootstrapTheme),
		page.Server(h.options.Proto+"://"+h.options.Addr),
		page.Version(h.options.Version),
		page.Totals(totals),
	)
	for _, d := range data {
		d(p)
	}

	e := p.Show(w)
	if e != nil {
		h.log.Logf("ERROR showError: failed to generate page: %v", e)
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

	h.showPage(w,
		page.Template("index.html"),
		page.Title(h.options.BrandName+" - Home"),
		page.Pastes(pastes),
		page.User(usr),
	)
}

// handlePostPaste creates new paste from the form data
func (h *Server) handlePostPaste(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)
	// Read the form data
	r.Body = http.MaxBytesReader(w, r.Body, h.options.MaxBodySize)
	if err := r.ParseForm(); err != nil {
		h.log.Logf("WARN parsing form failed: %v", err)
		h.showError(w, http.StatusBadRequest, "")
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
	var pr = service.PasteRequest{
		Title:           r.PostFormValue("title"),
		Body:            r.PostFormValue("body"),
		Expires:         r.PostFormValue("expires"),
		DeleteAfterRead: r.PostFormValue("delete_after_read") == "yes",
		Privacy:         r.PostFormValue("privacy"),
		Password:        r.PostFormValue("password"),
		Syntax:          r.PostFormValue("syntax"),
		UserID:          usr.ID,
	}
	paste, err := h.service.NewPaste(pr)
	if err != nil {
		if errors.Is(err, service.ErrEmptyBody) {
			h.showError(w, http.StatusBadRequest, "Body must not be empty.")
			return
		}
		if errors.Is(err, service.ErrWrongPrivacy) {
			h.showError(w, http.StatusBadRequest, "Privacy can be one of 'private', 'public' or 'unlisted'.")
			return
		}
		if errors.Is(err, service.ErrWrongDuration) {
			h.showError(w, http.StatusBadRequest, "Duration format is incorrect.")
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

	h.showPage(w,
		page.Template("view.html"),
		page.Title(h.options.BrandName+" - Paste"),
		page.Pastes(pastes),
		page.Paste(paste),
		page.User(usr),
	)
}

// handleGetPastePage generates a page to view a single paste.
func (h *Server) handleGetPastePage(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)
	// Get paste encoded ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		h.log.Logf("WARN handleGetPastePage: paste id not found")
		h.showError(w, http.StatusBadRequest, "")
		return
	}
	// If the request comes from a password form, get the password
	if err := r.ParseForm(); err != nil {
		h.log.Logf("WARN parsing form failed: %v", err)
		h.showError(w, http.StatusBadRequest, "")
		return
	}
	pwd := r.PostFormValue("password")

	// Get the paste from the storage
	paste, err := h.service.GetPaste(id, usr.ID, pwd)
	if err != nil {
		// Check if paste was not found
		if errors.Is(err, service.ErrPasteNotFound) {
			h.showError(w, http.StatusNotFound, "There is no such paste")
			return
		}
		// Check if paste is private an belongs to another user
		if errors.Is(err, service.ErrPasteIsPrivate) {
			h.showError(w, http.StatusForbidden, "This paste is private")
			return
		}
		// Check if paste is password-protected
		if errors.Is(err, service.ErrPasteHasPassword) || errors.Is(err, service.ErrWrongPassword) {
			w.WriteHeader(http.StatusUnauthorized)
			h.showPage(w,
				page.Template("password.html"),
				page.Title(h.options.BrandName+" - Password"),
				page.PasteID(id),
				page.User(usr),
				page.ErrorMessage("This paste is protected by a password"),
			)
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

	h.showPage(w,
		page.Template("view.html"),
		page.Title(h.options.BrandName+" - Paste"),
		page.Pastes(pastes),
		page.Paste(paste),
		page.User(usr),
	)
}

// handleGetPastesList generates a page to view a list of pastes.
func (h *Server) handleGetPastesList(w http.ResponseWriter, r *http.Request) {
	usr, _ := token.GetUserInfo(r)
	limit := 10 //TODO: make it configurable as PageSize or MaxPastesPerPage
	skip, err := strconv.Atoi(r.FormValue("skip"))
	if err != nil {
		skip = 0
	}

	pastes, err := h.service.GetPastes(usr.ID, "-created", limit, skip)
	if err != nil {
		h.showInternalError(w, err)
		return
	}
	count := h.service.PastesCount(usr.ID)                       // number of user pastes
	pageCount := int(math.Ceil(float64(count) / float64(limit))) // number of pages

	paginator := page.Paginator{
		Current:    skip/limit + 1,
		Last:       pageCount,
		LastOffset: (pageCount - 1) * limit,
		Pages:      make([]page.PaginatorLink, pageCount),
	}

	for i := 1; i <= pageCount; i++ {
		paginator.Pages[i-1] = page.PaginatorLink{
			Number: i,
			Offset: (i - 1) * limit,
		}
	}

	h.showPage(w,
		page.Template("list.html"),
		page.Title(h.options.BrandName+" - Pastes"),
		page.Pastes(pastes),
		page.PageLinks(paginator),
		page.User(usr),
	)
}

// Show 404 Not Found error page
func (h *Server) notFound(w http.ResponseWriter, r *http.Request) {
	h.showError(w, http.StatusNotFound, "Unfortunately the page you are looking for is not there 🙁")
}
