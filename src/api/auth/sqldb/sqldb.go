// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package sqldb provides an implementation of api.UserService that uses
// a database as a storage. It doesn't dictitate what type of database.
// Any database supported by the gorm(http://gorm.io) package will do.
package sqldb

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/auth"
	"gorm.io/gorm"
)

// DBUserStore is a database storage back-end that implements api.UserStore
// interface.
type DBUserStore struct {
	db *gorm.DB
}

// Store saves the user to the database.
func (s DBUserStore) Store(usr api.User) error {
	err := s.db.Create(&usr).Error
	if err != nil {
		return err
	}
	return nil
}

// Find searches for a user by one or several fields. Searchable fields are
// Username, Email and ID.
// For example, searching by username:
//	usr, err:= svc.Find(api.User{Username: "johnd"})
func (s DBUserStore) Find(usr api.User) (*api.User, error) {
	if s.db == nil {
		return nil, errors.New("Find: no database connection")
	}
	var u api.User
	err := s.db.Where(&usr).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Find: database error: %w", err)
	}

	return &u, nil
}

// SvcOptions contains all the options needed to create a new instance
// of DBUserService.
type SvcOptions struct {
	// Database connection string. See gorm package documentation for details.
	// For sqlite it can be either a file name or `file::memory:?cache=shared`
	// to use temporary database in memory (ex. for testing).
	DBConnection *gorm.DB
	// Whether to automatically create/update the database schema of service
	// creation. In any case, the auto migration will never delete or change
	// any data.
	DBAutoMigrate bool
}

// DBUserService is an implementation of the api.UserService interface with sql
// database as a storage.
type DBUserService struct {
	Options SvcOptions
	store   DBUserStore
	auth.UserService
}

// New initialises and returns an instance of UserService. It returns an error
// if there is a problem connecting to the database.
func New(opts SvcOptions) (*DBUserService, error) {
	var s DBUserService
	var err error
	s.Options = opts
	db := opts.DBConnection
	rand.Seed(time.Now().UnixNano())

	if s.Options.DBAutoMigrate {
		err = db.AutoMigrate(&api.User{})
	} else {
		if d, e := db.DB(); e == nil {
			err = d.Ping()
		} else {
			err = e
		}
	}
	s.store.db = db
	s.UserService.UserStore = s.store

	return &s, err
}
