// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package sqldb provides an implementation of api.UserService that uses
// a database as a storage.
package sqldb

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/iliafrenkel/go-pb/src/api"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SvcOptions contains all the options needed to create an instance
// of UserService
type SvcOptions struct {
	// Database connection string.
	// For sqlite it should be either a file name or `file::memory:?cache=shared`
	// to use temporary database in memory (ex. for testing).
	DBConnection *gorm.DB
	//
	DBAutoMigrate bool
	//
	TokenSecret string
}

// UserService stores all the users in sqlite database and implements
// auth.UserService interface.
type UserService struct {
	db      *gorm.DB
	Options SvcOptions
}

// New initialises and returns an instance of UserService.
func New(opts SvcOptions) (*UserService, error) {
	var s UserService
	s.Options = opts
	db := opts.DBConnection
	rand.Seed(time.Now().UnixNano())

	if s.Options.DBAutoMigrate {
		db.AutoMigrate(&api.User{})
	}
	s.db = db

	return &s, nil
}

// findByUsername finds a user by username.
// The return values are as follows:
// - if there is a problem talking to the database user == nil, err != nil
// - if user is not found user == nil, err == nil
// - if user is found user != nil, err == nil
func (s *UserService) findByUsername(uname string) (*api.User, error) {
	if s.db == nil {
		return nil, errors.New("findUserByName: no database connection")
	}
	var usr api.User
	err := s.db.Where("username = ?", uname).First(&usr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("findUserByName: database error: %w", err)
	}

	return &usr, nil
}

// findByEmail finds a user by email.
func (s *UserService) findByEmail(email string) (*api.User, error) {
	if s.db == nil {
		return nil, errors.New("findUserByName: no database connection")
	}
	var usr api.User
	err := s.db.Where("email = ?", email).First(&usr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("findUserByName: database error: %w", err)
	}

	return &usr, nil
}

// authToken returns an JWT token for provided user.
func (s *UserService) authToken(u api.User) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorised"] = true
	claims["user_id"] = u.ID
	claims["username"] = u.Username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	authToken, err := token.SignedString([]byte(s.Options.TokenSecret))
	return authToken, err
}

// Create creates a new user.
// Returns an error if user with the same username or the same email
// already exist or if passwords do not match.
func (s *UserService) Create(u api.UserRegister) error {
	usr, err := s.findByUsername(u.Username)
	if err != nil {
		return fmt.Errorf("Create: findByUsername failed: %w", err)
	}
	if usr != nil {
		return errors.New("user with such username already exists")
	}
	usr, err = s.findByEmail(u.Email)
	if err != nil {
		return fmt.Errorf("Create: findByEmail failed: %w", err)
	}
	if usr != nil {
		return errors.New("user with such email already exists")
	}
	if u.Password != u.RePassword {
		return errors.New("passwords don't match")
	}
	if u.Email == "" {
		return errors.New("email cannot be empty")
	}
	if u.Username == "" {
		return errors.New("username cannot be empty")
	}

	var newUsr api.User
	newUsr.ID = rand.Int63()
	newUsr.Username = u.Username
	newUsr.Email = strings.ToLower(u.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUsr.PasswordHash = string(hash)

	err = s.db.Create(&newUsr).Error
	if err != nil {
		return err
	}

	return nil
}

// Authenticate authenticates a user by validating that it exists and hash of the
// provided password matches. On success returns a JWT token.
// While this method returns different errors for different failures the
// end user should only see a generic "invalid credentials" message.
func (s *UserService) Authenticate(u api.UserLogin) (api.UserInfo, error) {
	inf := api.UserInfo{
		ID:       0,
		Username: "",
		Token:    "",
	}
	usr, err := s.findByUsername(u.Username)
	if err != nil {
		return inf, fmt.Errorf("Authenticate: findByUsername failed: %w", err)
	}
	if usr == nil {
		return inf, errors.New("user doesn't exist")
	}

	err = bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(u.Password))
	if err != nil {
		return inf, errors.New("invalid password")
	}

	token, err := s.authToken(*usr)

	if err != nil {
		return inf, err
	}

	inf.Username = usr.Username
	inf.Token = token
	return inf, nil
}

// Validate checks if provided token is valid. It returns auth.UserInfo with
// Username and Token if the token is valid or an empty UserInfo and an error
// if the token is invalid or if there was another error.
func (s *UserService) Validate(u api.User, t string) (api.UserInfo, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token signing method is not valid: %v", token.Header["alg"])
		}
		return []byte(s.Options.TokenSecret), nil
	})

	if err != nil {
		return api.UserInfo{}, err
	}

	var claims map[string]interface{}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		usr, err := s.findByUsername(claims["username"].(string))
		if err != nil {
			return api.UserInfo{}, fmt.Errorf("Validate: findByUsername failed: %w", err)
		}
		if usr != nil {
			return api.UserInfo{ID: usr.ID, Username: usr.Username, Token: token.Raw}, nil
		}
		return api.UserInfo{}, fmt.Errorf("token is valid but the user [%s] doesn't exist", claims["username"].(string))
	}
	return api.UserInfo{}, fmt.Errorf("alg header %v, error: %v", claims["alg"], err)
}
