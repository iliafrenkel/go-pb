// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.package main
package api

import (
	"time"

	"github.com/iliafrenkel/go-pb/src/api/base62"
)

// Paste is a the type that represents a single paste as it is stored in the
// database.
type Paste struct {
	ID              uint64    `json:"id"`
	Title           string    `json:"title" form:"title"`
	Body            string    `json:"body" form:"body" binding:"required"`
	Expires         time.Time `json:"expires"`
	DeleteAfterRead bool      `json:"delete_after_read" form:"delete_after_read" binding:"-"`
	Password        string    `json:"password"`
	Created         time.Time `json:"created"`
	Syntax          string    `json:"syntax" form:"syntax" binding:"required"`
	UserID          uint64    `json:"user_id"`
}

// URL generates a base62 encoded string from the ID.
func (p *Paste) URL() string {
	return base62.Encode(p.ID)
}

// PasteForm represents the data that we expect to recieve when the user
// submitts the form.
type PasteForm struct {
	Title           string `json:"title" form:"title"`
	Body            string `json:"body" form:"body" binding:"required"`
	Expires         string `json:"expires" form:"expires" binding:"required"`
	DeleteAfterRead bool   `json:"delete_after_read" form:"delete_after_read" binding:"-"`
	Password        string `json:"password" form:"password"`
	Syntax          string `json:"syntax" form:"syntax" binding:"required"`
	UserID          uint64 `json:"user_id"`
}

// PasteService is the interface that defines methods for working with Pastes.
//
// Implementations should define the underlying storage such as database,
// plain files or even memory.
type PasteService interface {
	Get(id uint64) (*Paste, error)
	Create(p PasteForm) (*Paste, error)
	Delete(id uint64) error
	List() []Paste
}

// User is a type that represents a single user as it is stored in the database
type User struct {
	ID           uint64 `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"_"`
}

// UserRegister represents the data that we expect to recieve from the
// registration page.
type UserRegister struct {
	Username   string `json:"username" form:"username" binding:"required"`
	Email      string `json:"email" form:"email" binding:"required"`
	Password   string `json:"password" form:"password" binding:"required"`
	RePassword string `json:"repassword" form:"repassword" binding:"required"`
}

// UserLogin represents the data that we expect to recieve from the
// login page.
type UserLogin struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// UserInfo represents the data that we send back in response to various
// operation such as Authenticate or Validate.
type UserInfo struct {
	ID       uint64 `json:"user_id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// UserService is the interface that defines methods to work with Users
type UserService interface {
	// Creates a new user.
	// Returns an error if user with the same username or the same email
	// already exist or if passwords do not match.
	Create(u UserRegister) error
	// Authenticates a user by validating that it exists and hash of the
	// provided password matches. On success returns a JWT token.
	Authenticate(u UserLogin) (UserInfo, error)
	// Validates given token for a given user.
	Validate(u User, t string) (UserInfo, error)
}
