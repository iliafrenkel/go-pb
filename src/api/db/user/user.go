package user

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"golang.org/x/crypto/bcrypt"
)

// UserService stores all the users in memory
type UserService struct {
	Users map[uint64]*api.User
}

// New creates a new UserService
func New() *UserService {
	var s UserService
	s.Users = make(map[uint64]*api.User)
	return &s
}

// findByUsername finds a user by username
func (s *UserService) findByUsername(uname string) *api.User {
	for _, u := range s.Users {
		if u.Username == uname {
			return u
		}
	}

	return nil
}

// findByEmail finds a user by email
func (s *UserService) findByEmail(email string) *api.User {
	for _, u := range s.Users {
		if u.Email == email {
			return u
		}
	}

	return nil
}

func (s *UserService) Create(usr api.User) (*api.User, error) {
	if s.findByUsername(usr.Username) != nil {
		return nil, errors.New("user with such username already exists")
	}
	if s.findByEmail(usr.Email) != nil {
		return nil, errors.New("user with such email already exists")
	}
	if usr.Password != usr.RePassword {
		return nil, errors.New("passwords don't match")
	}
	if usr.Email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if usr.Username == "" {
		return nil, errors.New("username cannot be empty")
	}

	rand.Seed(time.Now().UnixNano())
	usr.ID = rand.Uint64()
	usr.Email = strings.ToLower(usr.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	usr.PasswordHash = string(hash)
	usr.Password = ""
	usr.RePassword = ""

	s.Users[usr.ID] = &usr
	return &usr, nil
}

func (s *UserService) Authenticate(usr api.User) (string, error) {
	return "", nil
}
