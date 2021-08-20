// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package store defines a common interface that any concrete storage
// implementation must implement. Along wiht some supporting types.
// It provides two implementations of store.Interface - MemDB and PostgresDB.
package store

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Interface defines methods that an implementation of a concrete storage
// must provide.
type Interface interface {
	Count() (pastes, users int64)             // return total counts for pastes and users
	Create(paste Paste) (id int64, err error) // create new paste and return its id
	Delete(id int64) error                    // delete paste by id
	Find(req FindRequest) ([]Paste, error)    // find pastes
	Get(id int64) (Paste, error)              // get paste by id
	Update(paste Paste) (Paste, error)        // update paste information and return updated paste
	SaveUser(usr User) (id string, err error) // creates or updates a user
	User(id string) (User, error)             // get user by id
}

// User represents a single user.
type User struct {
	ID    string `json:"id" gorm:"primaryKey"`
	Name  string `json:"name" gorm:"index"`
	Email string `json:"email" gorm:"index"`
	IP    string `json:"ip,omitempty"`
	Admin bool   `json:"admin"`
}

// Paste represents a single paste with an optional reference to its user.
type Paste struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	Title           string    `json:"title"`
	Body            string    `json:"body"`
	Expires         time.Time `json:"expires" gorm:"index"`
	DeleteAfterRead bool      `json:"delete_after_read"`
	Privacy         string    `json:"privacy"`
	Password        string    `json:"password"`
	CreatedAt       time.Time `json:"created"`
	Syntax          string    `json:"syntax"`
	UserID          string    `json:"user_id" gorm:"index default:null"`
	User            User      `json:"user"`
	Views           int64     `json:"views"`
}

// URL generates a base62 encoded string from the paste ID. This string is
// used as a unique URL for the paste, hence the name.
func (p Paste) URL() string {
	const (
		alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length   = int64(len(alphabet))
	)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(11)
	number := p.ID

	for ; number > 0; number = number / length {
		encodedBuilder.WriteByte(alphabet[(number % length)])
	}

	return encodedBuilder.String()
}

// URL2ID decodes the previously generated URL string into a paste ID.
func (p Paste) URL2ID(url string) (int64, error) {
	const (
		alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length   = int64(len(alphabet))
	)

	var number int64

	for i, symbol := range url {
		alphabeticPosition := strings.IndexRune(alphabet, symbol)

		if alphabeticPosition == -1 {
			return int64(alphabeticPosition), errors.New("invalid character: " + string(symbol))
		}
		number += int64(alphabeticPosition) * int64(math.Pow(float64(length), float64(i)))
	}

	return number, nil
}

// Expiration returns a "humanized" duration between now and the expiry date
// stored in `Expires`. For example: "25 minutes" or "2 months" or "Never".
func (p Paste) Expiration() string {
	if p.Expires.IsZero() {
		return "Never"
	}

	diff := time.Time{}.Add(time.Until(p.Expires))
	years, months, days := diff.Date()
	hours, minutes, seconds := diff.Clock()

	switch {
	case years >= 2:
		return fmt.Sprintf("%d years", years-1)
	case months >= 2:
		return fmt.Sprintf("%d months", months-1)
	case days >= 2:
		return fmt.Sprintf("%d days", days-1+hours/12)
	case hours >= 1:
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	case minutes >= 2:
		return fmt.Sprintf("%d min", minutes)
	case seconds >= 1:
		return fmt.Sprintf("%d sec", seconds)
	}

	return p.Expires.Sub(p.CreatedAt).String()
}

// FindRequest is an input to the Find method
type FindRequest struct {
	UserID string
	Sort   string
	Since  time.Time
	Limit  int
	Skip   int
}
