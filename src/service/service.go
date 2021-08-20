// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package service provides methods to work with pastes and users.
// Methods of this package do not log or print out anything, they return
// errors instead. It is up to the user of the Service to handle the errors
// and provide useful information to the end user.
package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/iliafrenkel/go-pb/src/store"
	"golang.org/x/crypto/bcrypt"
)

// Service type provides method to work with pastes and users.
type Service struct {
	store store.Interface
}

// ErrPasteNotFound and other common errors.
var (
	ErrPasteNotFound    = errors.New("paste not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrPasteIsPrivate   = errors.New("paste is private")
	ErrPasteHasPassword = errors.New("paste has password")
	ErrWrongPassword    = errors.New("paste password is incorrect")
	ErrStoreFailure     = errors.New("store opertation failed")
	ErrEmptyBody        = errors.New("body is empty")
	ErrWrongPrivacy     = errors.New("privacy is wrong")
	ErrWrongDuration    = errors.New("wrong duration format")
)

// PasteRequest is an input to Create method, normally comes from a web form.
type PasteRequest struct {
	Title           string `json:"title" form:"title"`
	Body            string `json:"body" form:"body" binding:"required"`
	Expires         string `json:"expires" form:"expires" binding:"required"`
	DeleteAfterRead bool   `json:"delete_after_read" form:"delete_after_read" binding:"-"`
	Privacy         string `json:"privacy" form:"privacy" binding:"required"`
	Password        string `json:"password" form:"password"`
	Syntax          string `json:"syntax" form:"syntax" binding:"required"`
	UserID          string `json:"user_id"`
}

// New returns new Service with provided store as a back-end storage.
func New(store store.Interface) *Service {
	var s *Service = new(Service)
	s.store = store
	rand.Seed(time.Now().UnixNano())

	return s
}

// NewWithMemDB returns new Service with memory as a store.
func NewWithMemDB() *Service {
	return New(store.NewMemDB())
}

// NewWithPostgres returns new Service with postgres db as a store.
func NewWithPostgres(conn string) (*Service, error) {
	s, err := store.NewPostgresDB(conn, true)
	if err != nil {
		return nil, err
	}
	return New(s), nil
}

// parseExpiration tries to parse PasteRequest.Expires string and return
// corresponding time.Time.
// We expect the expiration to be in the form of "nx" where "n" is a number
// and "x" is a time unit character: m for minute, h for hour, d for day,
// w for week, M for month and y for year.
func (s Service) parseExpiration(exp string) (time.Time, error) {
	res := time.Time{}
	now := time.Now()

	if exp != "never" && len(exp) > 1 {
		dur, err := strconv.Atoi(exp[:len(exp)-1])
		if err != nil {
			return time.Time{}, fmt.Errorf("Service.parseExpiration: %w: %s (%v)", ErrWrongDuration, exp, err)
		}
		switch exp[len(exp)-1] {
		case 'm': //minutes
			res = now.Add(time.Duration(dur) * time.Minute)
		case 'h': //hours
			res = now.Add(time.Duration(dur) * time.Hour)
		case 'd': //days
			res = now.AddDate(0, 0, dur)
		case 'w': //weeks
			res = now.AddDate(0, 0, dur*7)
		case 'M': //months
			res = now.AddDate(0, dur, 0)
		case 'y': //years
			res = now.AddDate(dur, 0, 0)
		default:
			return time.Time{}, fmt.Errorf("Service.NewPaste: %w: %s", ErrWrongDuration, exp)
		}
	}
	return res, nil
}

// NewPaste creates new Paste from the request and saves it in the store.
// Paste.Body is mandatory, Paste.Expires is default to never, Paste.Privacy
// must be on of ["private","public","unlisted"]. If password is provided it
// is stored as a hash.
func (s Service) NewPaste(pr PasteRequest) (store.Paste, error) {
	var err error
	created := time.Now()
	expires, err := s.parseExpiration(pr.Expires)
	if err != nil {
		return store.Paste{}, fmt.Errorf("Service.NewPaste: %w", err)
	}

	// Check that body is not empty
	if pr.Body == "" {
		return store.Paste{}, ErrEmptyBody
	}

	// Privacy can only be "private", "public" or "unlisted"
	if pr.Privacy != "private" && pr.Privacy != "public" && pr.Privacy != "unlisted" {
		return store.Paste{}, ErrWrongPrivacy
	}

	// If password is not empty, hash it before storing
	if pr.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pr.Password), bcrypt.DefaultCost)
		if err != nil {
			return store.Paste{}, err
		}
		pr.Password = string(hash)
	}
	// If the user is known check that it is in our database and add if it's not
	var usr store.User
	if pr.UserID != "" {
		usr, err = s.store.User(pr.UserID)
		if err != nil || usr == (store.User{}) {
			return store.Paste{}, fmt.Errorf("Service.NewPaste: %w: user id [%s] (%v)", ErrUserNotFound, pr.UserID, err)
		}
	}
	// Default syntax to "text"
	if pr.Syntax == "" {
		pr.Syntax = "text"
	}
	// Create a new paste and store it
	paste := store.Paste{
		Title:           pr.Title,
		Body:            pr.Body,
		Expires:         expires,
		DeleteAfterRead: pr.DeleteAfterRead,
		Privacy:         pr.Privacy,
		Password:        pr.Password,
		CreatedAt:       created,
		Syntax:          pr.Syntax,
		User:            usr,
	}
	id, err := s.store.Create(paste)
	if err != nil {
		return store.Paste{}, fmt.Errorf("Service.NewPaste: %w: (%v)", ErrStoreFailure, err)
	}
	// Get the paste back and return it
	paste, err = s.store.Get(id)
	if err != nil {
		return store.Paste{}, fmt.Errorf("Service.NewPaste: %w: (%v)", ErrStoreFailure, err)
	}
	return paste, nil
}

// GetPaste returns a paste given encoded URL.
// If the paste is private GetPaste will check that it belongs to the user with
// provided uid. If password is given and the paste has password GetPaste will
// check that the password is correct.
func (s Service) GetPaste(url string, uid string, pwd string) (store.Paste, error) {
	p := store.Paste{}
	id, err := p.URL2ID(url)
	if err != nil {
		return store.Paste{}, err
	}
	p, err = s.store.Get(id)
	if err != nil {
		return p, fmt.Errorf("Service.GetPaste: %w: (%v)", ErrStoreFailure, err)
	}
	// Check if paste was not found
	if p == (store.Paste{}) {
		return p, fmt.Errorf("Service.GetPaste: %w: url [%s], id [%v]", ErrPasteNotFound, url, id)
	}
	// Check privacy
	if p.Privacy == "private" && p.User.ID != uid {
		return store.Paste{}, ErrPasteIsPrivate
	}
	// Check if password protected
	if p.Password != "" && pwd == "" {
		return store.Paste{}, ErrPasteHasPassword
	}
	// Check if password is correct
	if p.Password != "" && bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(pwd)) != nil {
		return store.Paste{}, ErrWrongPassword
	}
	// Update the view count
	p.Views++
	p, _ = s.store.Update(p) // we ignore the error here because we only update the view count
	// Check if paste is a "burner" and delete it if yes
	if p.DeleteAfterRead {
		err = s.store.Delete(p.ID)
		if err != nil {
			return p, fmt.Errorf("Service.GetPaste: %w: (%v)", ErrStoreFailure, err)
		}
	}
	return p, nil
}

// GetOrUpdateUser saves the user in the store and returns it.
func (s Service) GetOrUpdateUser(usr store.User) (store.User, error) {
	_, err := s.store.SaveUser(usr)
	if err != nil {
		return store.User{}, fmt.Errorf("Service.GetOrUpdateUser: %w: (%v)", ErrStoreFailure, err)
	}
	return usr, nil
}

// UserPastes returns a list of the last 10 paste for a user.
func (s Service) UserPastes(uid string) ([]store.Paste, error) {
	pastes, err := s.store.Find(store.FindRequest{
		UserID: uid,
		Sort:   "-created",
		Since:  time.Time{},
		Limit:  10,
		Skip:   0,
	})
	if err != nil {
		return nil, fmt.Errorf("Service.UserPastes: %w: (%v)", ErrStoreFailure, err)
	}
	return pastes, nil
}

// GetCount returns total count of pastes and users.
func (s Service) GetCount() (pastes, users int64) {
	return s.store.Count()
}
