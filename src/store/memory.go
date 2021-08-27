// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

package store

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
)

// MemDB is a memory storage that implements the store.Interface.
// Because it's a transient storage you will loose all the data once the
// process exits. It's not completely useless though. You can use it when a
// temporary sharing is needed or as a cache for another storage.
type MemDB struct {
	pastes map[int64]Paste
	users  map[string]User
	sync.RWMutex
}

// NewMemDB initialises and returns an instance of MemDB.
func NewMemDB() *MemDB {
	var s MemDB
	s.pastes = make(map[int64]Paste)
	s.users = make(map[string]User)

	return &s
}

// Count returns total count of pastes and users.
func (m *MemDB) Totals() (pastes, users int64) {
	m.RLock()
	defer m.RUnlock()

	return int64(len(m.pastes)), int64(len(m.users))
}

// Create creates and stores a new paste returning its ID.
func (m *MemDB) Create(p Paste) (id int64, err error) {
	m.Lock()
	defer m.Unlock()

	p.ID = rand.Int63() // #nosec
	m.pastes[p.ID] = p

	return p.ID, nil
}

// Delete deletes a paste by ID.
func (m *MemDB) Delete(id int64) error {
	m.Lock()
	defer m.Unlock()

	delete(m.pastes, id)

	return nil
}

// Find return a sorted list of pastes for a given request.
func (m *MemDB) Find(req FindRequest) (pastes []Paste, err error) {
	pastes = []Paste{}

	m.RLock()
	// Find all the pastes for a user
	for _, p := range m.pastes {
		if p.User.ID == req.UserID && p.CreatedAt.After(req.Since) {
			pastes = append(pastes, p)
		}
	}
	m.RUnlock()
	// Sort
	sort.Slice(pastes, func(i, j int) bool {
		switch req.Sort {
		case "+created", "-created":
			if strings.HasPrefix(req.Sort, "-") {
				return pastes[i].CreatedAt.After(pastes[j].CreatedAt)
			}
			return pastes[i].CreatedAt.Before(pastes[j].CreatedAt)
		case "+expires", "-expires":
			if strings.HasPrefix(req.Sort, "-") {
				return pastes[i].Expires.After(pastes[j].Expires)
			}
			return pastes[i].Expires.Before(pastes[j].Expires)
		case "+views", "-views":
			if strings.HasPrefix(req.Sort, "-") {
				return pastes[i].Views > pastes[j].Views
			}
			return pastes[i].Views <= pastes[j].Views
		default:
			return pastes[i].CreatedAt.Before(pastes[j].CreatedAt)
		}
	})
	// Slice with skip and limit
	skip := req.Skip
	if skip > len(pastes) {
		skip = len(pastes)
	}
	end := skip + req.Limit
	if end > len(pastes) {
		end = len(pastes)
	}

	return pastes[skip:end], nil
}

func (m *MemDB) Count(uid string) int64 {
	m.RLock()
	defer m.RUnlock()
	// Count all the pastes for a user
	var cnt int64
	for _, p := range m.pastes {
		if p.User.ID == uid {
			cnt++
		}
	}
	return cnt
}

// Get returns a paste by ID.
func (m *MemDB) Get(id int64) (Paste, error) {
	m.RLock()
	defer m.RUnlock()

	return m.pastes[id], nil
}

// SaveUser creates a new or updates an existing user.
func (m *MemDB) SaveUser(usr User) (id string, err error) {
	m.Lock()
	defer m.Unlock()

	m.users[usr.ID] = usr

	return usr.ID, nil
}

// User returns a user by ID.
func (m *MemDB) User(id string) (User, error) {
	m.RLock()
	defer m.RUnlock()

	var usr User
	var ok bool

	if usr, ok = m.users[id]; !ok {
		return User{}, fmt.Errorf("MemDB.User: user not found")
	}
	return usr, nil
}

// Update updates existing paste.
func (m *MemDB) Update(p Paste) (Paste, error) {
	m.RLock()
	if _, ok := m.pastes[p.ID]; !ok {
		m.RUnlock()
		return Paste{}, nil
	}
	m.RUnlock()
	m.Lock()
	defer m.Unlock()
	m.pastes[p.ID] = p
	return p, nil
}
