// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

package paste

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"golang.org/x/crypto/bcrypt"
)

type PasteService struct {
	PasteStore api.PasteStore
}

// Get returns a paste by ID. If the paste is not found no error is returned.
// Instead, both return values are nil.
func (s *PasteService) Get(id int64) (*api.Paste, error) {
	pastes, err := s.PasteStore.Find(api.Paste{ID: id})
	if err != nil {
		return nil, err
	}
	if len(pastes) == 0 {
		return nil, nil
	}
	return &pastes[0], nil
}

// Create creates new paste and returns it on success.
func (s *PasteService) Create(p api.PasteForm) (*api.Paste, error) {
	var (
		expires, created time.Time
	)
	created = time.Now()
	expires = time.Time{} // zero time means no expiration, this is the default
	// We expect the expiration to be in the form of "nx" where "n" is a number
	// and "x" is a time unit character: m for minute, h for hour, d for day,
	// w for week, M for month and y for year.
	if p.Expires != "never" && len(p.Expires) > 1 {
		dur, err := strconv.Atoi(p.Expires[:len(p.Expires)-1])
		if err != nil {
			return nil, fmt.Errorf("wrong duration format: %s, error: %w", p.Expires, err)
		}
		switch p.Expires[len(p.Expires)-1] {
		case 'm': //minutes
			expires = created.Add(time.Duration(dur) * time.Minute)
		case 'h': //hours
			expires = created.Add(time.Duration(dur) * time.Hour)
		case 'd': //days
			expires = created.AddDate(0, 0, dur)
		case 'w': //weeks
			expires = created.AddDate(0, 0, dur*7)
		case 'M': //months
			expires = created.AddDate(0, dur, 0)
		case 'y': //days
			expires = created.AddDate(dur, 0, 0)
		default:
			return nil, fmt.Errorf("unknown duration format: %s", p.Expires)
		}
	}
	// Create new paste with a randomly generated ID and a hashed password.
	if p.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		p.Password = string(hash)
	}
	newPaste := api.Paste{
		ID:              rand.Int63(),
		Title:           p.Title,
		Body:            p.Body,
		Expires:         expires,
		DeleteAfterRead: p.DeleteAfterRead,
		Privacy:         p.Privacy,
		Password:        p.Password,
		Created:         created,
		Syntax:          p.Syntax,
		UserID:          p.UserID,
	}

	if err := s.PasteStore.Store(newPaste); err != nil {
		return nil, err
	}

	return &newPaste, nil
}

// Delete deletes a paste by ID. If paste with the provided ID doesn't
// exist this method does nothing, it will not return an error in such case.
func (s *PasteService) Delete(id int64) error {
	return s.PasteStore.Delete(api.Paste{ID: id})
}

// List returns a list of pastes for a user specified by ID.
func (s *PasteService) List(uid int64) ([]api.Paste, error) {
	pastes, err := s.PasteStore.Find(api.Paste{UserID: uid})
	if err != nil {
		return nil, err
	}
	return pastes, nil
}
