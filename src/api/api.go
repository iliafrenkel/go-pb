// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package api is an entry point to go-pb API. It defines all the types
// and interfaces that needed to implemented.
package api

import (
	"fmt"
	"time"

	"github.com/iliafrenkel/go-pb/src/api/base62"
)

// Paste is a the type that represents a single paste as it is stored in the
// database.
type Paste struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	Title           string    `json:"title"`
	Body            string    `json:"body"`
	Expires         time.Time `json:"expires" gorm:"index"`
	DeleteAfterRead bool      `json:"delete_after_read"`
	Password        string    `json:"password"`
	Created         time.Time `json:"created"`
	Syntax          string    `json:"syntax"`
	UserID          int64     `json:"user_id" gorm:"index"`
	User            User      `json:"-"`
}

// URL generates a base62 encoded string from the ID.
func (p *Paste) URL() string {
	return base62.Encode(uint64(p.ID))
}

// Expiration returns a "humanized" duration between now and the expiry date
// stored in `Expires`. For example: "in 25 minutes" or "in 2 months".
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

	lbl := "in"

	switch {
	case diff <= 0:
		return "now"
	case diff <= 2:
		return fmt.Sprintf("%s 1 second", lbl)
	case diff < 1*Minute:
		return fmt.Sprintf("%s %d seconds", lbl, diff)

	case diff < 2*Minute:
		return fmt.Sprintf("%s 1 minute", lbl)
	case diff < 1*Hour:
		return fmt.Sprintf("%s %d minutes", lbl, diff/Minute)

	case diff < 2*Hour:
		return fmt.Sprintf("%s 1 hour", lbl)
	case diff < 1*Day:
		return fmt.Sprintf("%s %d hours", lbl, diff/Hour)

	case diff < 2*Day:
		return fmt.Sprintf("%s 1 day", lbl)
	case diff < 1*Week:
		return fmt.Sprintf("%s %d days", lbl, diff/Day)

	case diff < 2*Week:
		return fmt.Sprintf("%s 1 week", lbl)
	case diff < 1*Month:
		return fmt.Sprintf("%s %d weeks", lbl, diff/Week)

	case diff < 2*Month:
		return fmt.Sprintf("%s 1 month", lbl)
	case diff < 1*Year:
		return fmt.Sprintf("%s %d months", lbl, diff/Month)

	case diff < 18*Month:
		return fmt.Sprintf("%s ~1 year", lbl)
	case diff < 2*Year:
		return fmt.Sprintf("%s ~2 years", lbl)
	case diff < LongTime:
		return fmt.Sprintf("%s %d years", lbl, diff/Year)
	}

	return p.Expires.Sub(p.Created).String()
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
	UserID          int64  `json:"user_id"`
}

// PasteService is the interface that defines methods for working with Pastes.
//
// Implementations should define the underlying storage such as database,
// plain files or even memory.
type PasteService interface {
	// Get returns a paste by ID.
	// Please note that error should be not nil only if there is an actual
	// error. Paste not found is not an error. It should be expected that if
	// there is no paste with the ID that you provided the return value for
	// the paste will be nil.
	Get(id int64) (*Paste, error)
	// Create creates new paste, saves it to the storage and returns it.
	Create(p PasteForm) (*Paste, error)
	// Delete deletes a paste by ID.
	// If paste with the provided ID doesn't exist this method does nothing,
	// it will not return an error in such case.
	Delete(id int64) error
	// List returns all the pastes with specified user ID.
	List(uid int64) []Paste
}

// User is a type that represents a single user as it is stored in the database
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	Username     string    `json:"username" gorm:"index"`
	Email        string    `json:"email" gorm:"index"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
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
