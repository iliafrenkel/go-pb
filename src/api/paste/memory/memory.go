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
	"math/rand"
	"sync"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/paste"
)

// MemoryPasteStore is a thread safe memory storage that implements
// api.PasteStore interface.
type MemoryPasteStore struct {
	pastes     map[int64]*api.Paste
	pastesLock *sync.RWMutex // controls access to pastes map
}

// Store adds the user to the internal map struct.
func (s MemoryPasteStore) Store(paste api.Paste) error {
	s.pastesLock.Lock()
	defer s.pastesLock.Unlock()
	s.pastes[paste.ID] = &paste

	return nil
}

// Find searches for a pastes by one or several fields. Searchable fields are
// ID, Expires and UserID.
// For example, searching by ID:
//	pastes, err := svc.Find(api.Paste{ID: 12345})
func (s MemoryPasteStore) Find(paste api.Paste) ([]api.Paste, error) {
	pastes := make([]api.Paste, 0, len(s.pastes))
	s.pastesLock.RLock()
	defer s.pastesLock.RUnlock()
	for _, p := range s.pastes {
		if paste.ID != 0 && paste.ID == p.ID {
			pastes = append(pastes, *p)
		} else if !paste.Expires.IsZero() && paste.Expires == p.Expires {
			pastes = append(pastes, *p)
		} else if paste.UserID != 0 && paste.UserID == p.UserID {
			pastes = append(pastes, *p)
		}
	}

	return pastes, nil
}

// Delete removes a paste from the map. The paste is identified by one or
// several fields. Searchable fields are ID, Expires and UserID.
// For example, delete by ID:
//	err := svc.Delete(api.Paste{ID: 12345})
func (s MemoryPasteStore) Delete(paste api.Paste) error {
	var id int64
	s.pastesLock.RLock()
	for _, p := range s.pastes {
		if paste.ID != 0 && paste.ID == p.ID {
			id = p.ID
		}
		if !paste.Expires.IsZero() && paste.Expires == p.Expires {
			id = p.ID
		}
		if paste.UserID != 0 && paste.UserID == p.UserID {
			id = p.ID
		}
	}
	s.pastesLock.RUnlock()

	s.pastesLock.Lock()
	defer s.pastesLock.Unlock()
	delete(s.pastes, id)

	return nil
}

// PasteService stores all the pastes in memory and implements the
// api.PasteService interface.
type PasteService struct {
	store MemoryPasteStore
	paste.PasteService
}

// New returns new PasteService with an empty map of pastes.
func New() *PasteService {
	var s PasteService
	s.store.pastes = make(map[int64]*api.Paste)
	s.store.pastesLock = &sync.RWMutex{}
	s.PasteService.PasteStore = s.store
	rand.Seed(time.Now().UnixNano())
	return &s
}
