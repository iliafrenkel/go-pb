// Copyright 2021 Ilia Frenkel. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.txt file.

// Package auth contains services for user management, authentication,
// and authorisation.
package auth

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/iliafrenkel/go-pb/src/api"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements api.UserService interface. This particular
// implementation doesn't have a storage back-end. It's main purpose is to
// be included within other implementations. See memory.UserService and
// sqldb.UserService for examples.
type UserService struct {
	UserStore api.UserStore
}

// Create recieves user registration information and creates a new user.
// It returns an error if the information doesn't pass validation or if it
// fails to store the new user.
func (s UserService) Create(usr api.UserRegister) error {
	// Validation: check that username/email is available and not empty,
	// and that passwords match.
	u, err := s.UserStore.Find(api.User{Username: usr.Username})
	if err != nil {
		return fmt.Errorf("Create: findByUsername failed: %w", err)
	}
	if u != nil {
		return errors.New("user with such username already exists")
	}
	u, err = s.UserStore.Find(api.User{Email: usr.Email})
	if err != nil {
		return fmt.Errorf("Create: findByEmail failed: %w", err)
	}
	if u != nil {
		return errors.New("user with such email already exists")
	}
	if usr.Password != usr.RePassword {
		return errors.New("passwords don't match")
	}
	if usr.Email == "" {
		return errors.New("email cannot be empty")
	}
	if usr.Username == "" {
		return errors.New("username cannot be empty")
	}
	// Create new user
	var newUsr api.User
	newUsr.ID = rand.Int63()
	newUsr.Username = usr.Username
	newUsr.Email = strings.ToLower(usr.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	newUsr.PasswordHash = string(hash)

	return s.UserStore.Store(newUsr)
}

// Authenticate authenticates a user by validating that it exists and that the
// hash of the provided password matches. On success it returns a JWT token.
// While this method returns different errors for different failures the
// end user should only be shown a generic "invalid credentials" message.
func (s UserService) Authenticate(usr api.UserLogin, secret string) (api.UserInfo, error) {
	inf := api.UserInfo{}
	// Find the user
	u, err := s.UserStore.Find(api.User{Username: usr.Username})
	if err != nil {
		return inf, fmt.Errorf("Authenticate: findByUsername failed: %w", err)
	}
	if u == nil {
		return inf, errors.New("user doesn't exist")
	}
	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(usr.Password))
	if err != nil {
		return inf, errors.New("invalid password")
	}
	// Generate JWT token
	token, err := u.GenerateAuthToken(secret)
	if err != nil {
		return inf, err
	}

	inf.Username = u.Username
	inf.Token = token
	return inf, nil
}

// Validate checks if provided token is valid. It returns auth.UserInfo with
// Username and Token if the token is valid or an empty UserInfo and an error
// if the token is invalid or if there was another error.
func (s UserService) Validate(usr api.User, token string, secret string) (api.UserInfo, error) {
	tkn, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token signing method is not valid: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return api.UserInfo{}, err
	}

	var claims map[string]interface{}
	if claims, ok := tkn.Claims.(jwt.MapClaims); ok && tkn.Valid {
		u, err := s.UserStore.Find(api.User{Username: claims["username"].(string)})
		if err != nil {
			return api.UserInfo{}, fmt.Errorf("Validate: findByUsername failed: %w", err)
		}
		if u != nil {
			return api.UserInfo{ID: u.ID, Username: u.Username, Token: tkn.Raw}, nil
		}
		return api.UserInfo{}, fmt.Errorf("token is valid but the user [%s] doesn't exist", claims["username"].(string))
	}
	return api.UserInfo{}, fmt.Errorf("alg header %v, error: %v", claims["alg"], err)
}
