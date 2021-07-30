// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package api defines types and interfaces used to implement go-pb API.
package api

import (
	"fmt"
	"time"

	"github.com/iliafrenkel/go-pb/src/api/base62"
)

// Paste represents a single paste object the way it is stored in the database.
type Paste struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	Title           string    `json:"title"`
	Body            string    `json:"body"`
	Expires         time.Time `json:"expires" gorm:"index"`
	DeleteAfterRead bool      `json:"delete_after_read"`
	Privacy         string    `json:"privacy"`
	Password        string    `json:"password"`
	Created         time.Time `json:"created"`
	Syntax          string    `json:"syntax"`
	UserID          int64     `json:"user_id" gorm:"index default:null"`
	User            User      `json:"-"`
}

// URL generates a base62 encoded string from the paste ID. This string is
// used as a unique URL for the paste, hence the name.
func (p *Paste) URL() string {
	return base62.Encode(p.ID)
}

// Expiration returns a "humanized" duration between now and the expiry date
// stored in `Expires`. For example: "25 minutes" or "2 months".
func (p *Paste) Expiration() string {
	if p.Expires.IsZero() {
		return "Never"
	}
	// Seconds-based time units
	const (
		Minute   = 60
		Hour     = 60 * Minute
		Day      = 24 * Hour
		Week     = 7 * Day
		Month    = 30 * Day
		Year     = 12 * Month
		LongTime = 37 * Year
	)

	diff := time.Until(p.Expires) / time.Second

	switch {
	case diff <= 0:
		return "now"
	case diff <= 2:
		return "1 second"
	case diff < 1*Minute:
		return fmt.Sprintf("%d seconds", diff)

	case diff < 2*Minute:
		return "1 minute"
	case diff < 1*Hour:
		return fmt.Sprintf("%d minutes", diff/Minute)

	case diff < 2*Hour:
		return "1 hour"
	case diff < 1*Day:
		return fmt.Sprintf("%d hours", diff/Hour)

	case diff < 2*Day:
		return "1 day"
	case diff < 1*Week:
		return fmt.Sprintf("%d days", diff/Day)

	case diff < 2*Week:
		return "1 week"
	case diff < 1*Month:
		return fmt.Sprintf("%d weeks", diff/Week)

	case diff < 2*Month:
		return "1 month"
	case diff < 1*Year:
		return fmt.Sprintf("%d months", diff/Month)

	case diff < 18*Month:
		return "~1 year"
	case diff < 2*Year:
		return "~2 years"
	case diff < LongTime:
		return fmt.Sprintf("%d years", diff/Year)
	}

	return p.Expires.Sub(p.Created).String()
}

// PasteForm represents the data that the PasteService.Create method expects.
// The data normally comes from a web form.
type PasteForm struct {
	Title           string `json:"title" form:"title"`
	Body            string `json:"body" form:"body" binding:"required"`
	Expires         string `json:"expires" form:"expires" binding:"required"`
	DeleteAfterRead bool   `json:"delete_after_read" form:"delete_after_read" binding:"-"`
	Privacy         string `json:"privacy" form:"privacy" binding:"required"`
	Password        string `json:"password" form:"password"`
	Syntax          string `json:"syntax" form:"syntax" binding:"required"`
	UserID          int64  `json:"user_id"`
}

// PastePassword represents the data that the API expects to verify the paste
// password. Event though it only has one field, we still use a struct for
// consistency. It makes it easier for the http package to implement common
// payload verification.
type PastePassword struct {
	Password string `json:"password" form:"password" binding:"required"`
}

// PasteService is the interface that defines methods to work with Pastes.
// Various implementations of this interface may use different storage
// mechanisms such as sql database, memory or plain files.
type PasteService interface {
	// Get returns a paste by ID. If the paste is not found no error is
	// returned. Instead, both return values are nil.
	Get(id int64) (*Paste, error)
	// Create creates new paste and returns it on success.
	Create(p PasteForm) (*Paste, error)
	// Delete deletes a paste by ID. If paste with the provided ID doesn't
	// exist this method does nothing, it will not return an error in such case.
	Delete(id int64) error
	// List returns a list of pastes for a user specified by ID.
	List(uid int64) []Paste
}

// User is a type that represents a single user the way it is stored in the
// database.
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"index"`
	Email        string    `json:"email" gorm:"index"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
}

// UserRegister represents the data that we expect to recieve from the
// registration form.
type UserRegister struct {
	Username   string `json:"username" form:"username" binding:"required"`
	Email      string `json:"email" form:"email" binding:"required"`
	Password   string `json:"password" form:"password" binding:"required"`
	RePassword string `json:"repassword" form:"repassword" binding:"required"`
}

// UserLogin represents the data that we expect to recieve from the
// login form.
type UserLogin struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// UserInfo represents the data that we send back in response to various
// operation such as Authenticate or Validate.
type UserInfo struct {
	ID       int64  `json:"user_id"`
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

// HTTPError represents an error that API sends to consumers.
type HTTPError struct {
	Code    int    `json:"code"`    // HTTP status code
	Message string `json:"message"` // Error message
}
