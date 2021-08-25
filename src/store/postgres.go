// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

package store

import (
	"fmt"
	"math/rand"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresDB is a Postgres SQL databasse storage that implements the
// store.Interface.
type PostgresDB struct {
	db *gorm.DB
}

// NewPostgresDB initialises a new instance of PostgresDB and returns.
// It tries to establish a database connection specified by conn and if
// autoMigrate is true it will try and create/alter all the tables.
func NewPostgresDB(conn string, autoMigrate bool) (*PostgresDB, error) {
	var pg PostgresDB
	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("NewPostgresDB: failed to establish database connection: %w", err)
	}
	if autoMigrate {
		err = db.AutoMigrate(&Paste{})
	} else {
		if d, e := db.DB(); e == nil {
			err = d.Ping()
		} else {
			err = e
		}
	}

	if err != nil {
		return nil, fmt.Errorf("NewPostgresDB: %w", err)
	}

	pg.db = db

	return &pg, nil
}

// Count returns total count of pastes and users.
func (pg *PostgresDB) Totals() (pastes, users int64) {
	pg.db.Model(&Paste{}).Count(&pastes)
	pg.db.Model(&User{}).Count(&users)
	return
}

// Create creates and stores a new paste returning its ID.
func (pg *PostgresDB) Create(p Paste) (id int64, err error) {
	p.ID = rand.Int63() // #nosec
	if p.User.ID == "" {
		err = pg.db.Omit("user_id").Create(&p).Error
	} else {
		err = pg.db.Create(&p).Error
	}
	if err != nil {
		return 0, fmt.Errorf("PostgresDB.Create: %w", err)
	}
	return p.ID, nil
}

// Delete deletes a paste by ID.
func (pg *PostgresDB) Delete(id int64) error {
	if id == 0 {
		return fmt.Errorf("PostgresDB.Delete: id cannot be null")
	}
	tx := pg.db.Delete(&Paste{}, id)
	err := tx.Error
	if err != nil {
		return fmt.Errorf("PostgresDB.Delete: %w", err)
	}
	if tx.RowsAffected == 0 {
		return fmt.Errorf("PostgresDB.Delete: no rows deleted")
	}

	return nil
}

// Find return a sorted list of pastes for a given request.
func (pg *PostgresDB) Find(req FindRequest) (pastes []Paste, err error) {
	sort := "created_at desc"
	switch req.Sort {
	case "+created", "-created":
		sort = "created_at"
		if strings.HasPrefix(req.Sort, "-") {
			sort = "created_at desc"
		}
	case "+expires", "-expires":
		sort = "expires"
		if strings.HasPrefix(req.Sort, "-") {
			sort = "expires desc"
		}
	case "+views", "-views":
		sort = "views"
		if strings.HasPrefix(req.Sort, "-") {
			sort = "views desc"
		}
	}

	err = pg.db.
		Limit(req.Limit).
		Offset(req.Skip).
		Order(sort).
		Select("id", "title", "expires", "delete_after_read", "privacy", "password", "created_at", "syntax", "views").
		Find(&pastes, "user_id = ?", req.UserID).Error
	if err != nil {
		return pastes, fmt.Errorf("PostgresDB.Find: %w", err)
	}
	return pastes, nil
}

func (pg *PostgresDB) Count(uid string) (pastes int64) {
	pg.db.Model(&Paste{}).Count(&pastes).Where("user_id = ?", uid)
	fmt.Printf("user: %s, count: %d", uid, pastes)
	return
}

// Get returns a paste by ID.
func (pg *PostgresDB) Get(id int64) (Paste, error) {
	var paste Paste
	err := pg.db.Preload("User").Limit(1).Find(&paste, id).Error
	if err != nil {
		return paste, fmt.Errorf("PostgresDB.Get: %w", err)
	}

	return paste, nil
}

// SaveUser creates a new or updates an existing user.
func (pg *PostgresDB) SaveUser(usr User) (id string, err error) {
	err = pg.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(&usr).Error
	if err != nil {
		return "", fmt.Errorf("PostgresDB.SaveUser: %w", err)
	}
	id = usr.ID
	return id, nil
}

// User returns a user by ID.
func (pg *PostgresDB) User(id string) (User, error) {
	var usr User
	tx := pg.db.Limit(1).Find(&usr, User{ID: id})
	err := tx.Error
	if err != nil {
		return usr, fmt.Errorf("PostgresDB.User: %w", err)
	}
	if tx.RowsAffected == 0 {
		return usr, fmt.Errorf("PostgresDB.User: user not found")
	}

	return usr, err
}

// Update saves the paste into database and returns it
func (pg *PostgresDB) Update(p Paste) (Paste, error) {
	err := pg.db.First(&Paste{}, p.ID).Error
	if err != nil {
		return Paste{}, fmt.Errorf("PostgresDB.Update: %w", err)
	}
	err = pg.db.Save(&p).Error
	if err != nil {
		return Paste{}, fmt.Errorf("PostgresDB.Update: %w", err)
	}

	return p, nil
}
