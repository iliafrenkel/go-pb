// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package sqldb provides implementation of api.PasteService that uses
// a database as a storage.
package sqldb

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/paste"
	"gorm.io/gorm"
)

// DBPasteStore is a database storage back-end that implements
// api.PasteStore interface.
type DBPasteStore struct {
	db *gorm.DB
}

// Store adds the user to the internal map struct.
func (s DBPasteStore) Store(paste api.Paste) error {
	var err error
	if paste.UserID == 0 {
		err = s.db.Omit("user_id").Create(&paste).Error
	} else {
		err = s.db.Create(&paste).Error
	}
	if err != nil {
		return err
	}

	return nil
}

// Find searches for a pastes by one or several fields. Searchable fields are
// ID, Expires and UserID.
// For example, searching by ID:
//	pastes, err := svc.Find(api.Paste{ID: 12345})
func (s DBPasteStore) Find(paste api.Paste) ([]api.Paste, error) {
	if s.db == nil {
		return nil, errors.New("Find: no database connection")
	}
	var pastes []api.Paste

	err := s.db.Where(&paste).Find(&pastes).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Find: database error: %w", err)
	}

	return pastes, nil
}

// Delete removes a paste from the map. The paste is identified by one or
// several fields. Searchable fields are ID, Expires and UserID.
// For example, delete by ID:
//	err := svc.Delete(api.Paste{ID: 12345})
func (s DBPasteStore) Delete(paste api.Paste) error {
	return s.db.Delete(&api.Paste{}, &paste).Error
}

// SvcOptions contains all the options needed to create an instance
// of PasteService
type SvcOptions struct {
	// Database connection string.
	// For sqlite it should be either a file name or `file::memory:?cache=shared`
	// to use temporary database in memory (ex. for testing).
	DBConnection *gorm.DB
	//
	DBAutoMigrate bool
}

// PasteService stores all the pastes in a database and implements the
// api.PasteService interface.
type DBPasteService struct {
	Options SvcOptions
	store   DBPasteStore
	paste.PasteService
}

// New returns new PasteService with an empty map of pastes.
func New(opts SvcOptions) (*DBPasteService, error) {
	var s DBPasteService
	var err error
	s.Options = opts
	db := opts.DBConnection
	rand.Seed(time.Now().UnixNano())

	if s.Options.DBAutoMigrate {
		db.AutoMigrate(&api.Paste{})
	} else {
		if d, e := db.DB(); e == nil {
			err = d.Ping()
		} else {
			err = e
		}
	}
	s.store.db = db
	s.PasteService.PasteStore = s.store

	return &s, err
}
