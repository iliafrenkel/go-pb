// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package memory provides methods to work with pastes using memory as a
// storage.
//
// This package provides a PasteService type that implements api.PasteService
// interface and use a map of Pastes as a storage.
//
// Note: according to the documentation (https://blog.golang.org/maps#TOC_6.),
// maps are not safe for concurrent use.
package memory

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
)

// PasteService stores all the pastes in memory and implements the
// api.PasteService interface.
type PasteService struct {
	pastes     map[int64]*api.Paste
	pastesLock *sync.RWMutex // controls access to pastes map
}

// New returns new PasteService with an empty map of pastes.
func New() *PasteService {
	var s PasteService
	s.pastes = make(map[int64]*api.Paste)
	s.pastesLock = &sync.RWMutex{}
	return &s
}

// Get returns a paste by it's ID.
func (s *PasteService) Get(id int64) (*api.Paste, error) {
	s.pastesLock.RLock()
	defer s.pastesLock.RUnlock()
	if p, ok := s.pastes[id]; ok {
		return p, nil
	}
	return nil, nil
}

// Create initialises a new paste from the provided data and adds it to the
// storage. It returns the newly created paste.
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
	// Create new paste with a randomly generated ID
	rand.Seed(time.Now().UnixNano())
	newPaste := api.Paste{
		ID:              rand.Int63(),
		Title:           p.Title,
		Body:            p.Body,
		Expires:         expires,
		DeleteAfterRead: p.DeleteAfterRead,
		Password:        p.Password,
		Created:         created,
		Syntax:          p.Syntax,
		UserID:          p.UserID,
	}

	s.pastesLock.Lock()
	defer s.pastesLock.Unlock()
	s.pastes[newPaste.ID] = &newPaste

	return &newPaste, nil
}

// Delete removes the paste from the storage
func (s *PasteService) Delete(id int64) error {
	s.pastesLock.RLock()
	if _, ok := s.pastes[id]; !ok {
		s.pastesLock.RUnlock()
		return nil
	}
	s.pastesLock.RUnlock()

	s.pastesLock.Lock()
	defer s.pastesLock.Unlock()
	delete(s.pastes, id)

	return nil
}

// List returns a slice of all the pastes by a user ID.
//
// TODO: remove the second condition (id == 0), it's a hack for now to list
// all the pastes.
func (s *PasteService) List(uid int64) []api.Paste {
	pastes := make([]api.Paste, 0, len(s.pastes))
	s.pastesLock.RLock()
	defer s.pastesLock.RUnlock()
	for _, val := range s.pastes {
		if val.UserID == uid || uid == 0 {
			pastes = append(pastes, *val)
		}
	}
	return pastes
}
