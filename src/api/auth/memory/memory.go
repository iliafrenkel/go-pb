// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package memory provides an implementation of api.UserService that uses
// memory as a storage.
package memory

import (
	"math/rand"
	"sync"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/auth"
)

// MemoryUserStore is a thread safe memory storage that implements
// api.UserStore interface.
type MemoryUserStore struct {
	users     map[int64]*api.User
	usersLock *sync.RWMutex // controls access to users map
}

// Store adds the user to the internal map struct.
func (s MemoryUserStore) Store(usr api.User) error {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	s.users[usr.ID] = &usr

	return nil
}

// Find searches for a user by one or several fields. Searchable fields are
// Username, Email and ID.
// For example, searching by username:
//	usr, err:= svc.Find(api.User{Username: "johnd"})
func (s MemoryUserStore) Find(usr api.User) (*api.User, error) {
	s.usersLock.RLock()
	defer s.usersLock.RUnlock()
	for _, u := range s.users {
		if usr.Username != "" && u.Username == usr.Username {
			return u, nil
		}
		if usr.Email != "" && u.Email == usr.Email {
			return u, nil
		}
		if usr.ID != 0 && u.ID == usr.ID {
			return u, nil
		}
	}

	return nil, nil
}

// MemoryUserService is an implementation of api.UserService with
// MemoryUserStore as a back-end storage.
type MemoryUserService struct {
	store MemoryUserStore
	auth.UserService
}

// New returns a new UserService.
// It initialises the underlying storage which in this case is map.
func New() *MemoryUserService {
	var s MemoryUserService
	s.store.users = make(map[int64]*api.User)
	s.store.usersLock = &sync.RWMutex{}
	s.UserService.UserStore = s.store
	rand.Seed(time.Now().UnixNano())
	return &s
}
