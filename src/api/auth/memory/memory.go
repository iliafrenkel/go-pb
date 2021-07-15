package memory

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/iliafrenkel/go-pb/src/api/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	tokenSecret = []byte("hardcodeddefault") // TODO:(os.Getenv("GOPB_TOKEN_SECRET"))
)

// UserService stores all the users in memory and implements auth.UserService
// interface.
type UserService struct {
	Users map[uint64]*auth.User
}

// New returns a new UserService.
// It initialises the underlying storage which in this case is map.
func New() *UserService {
	var s UserService
	s.Users = make(map[uint64]*auth.User)
	return &s
}

// findByUsername finds a user by username.
func (s *UserService) findByUsername(uname string) *auth.User {
	for _, u := range s.Users {
		if u.Username == uname {
			return u
		}
	}

	return nil
}

// findByEmail finds a user by email.
func (s *UserService) findByEmail(email string) *auth.User {
	for _, u := range s.Users {
		if u.Email == email {
			return u
		}
	}

	return nil
}

// authToken returns an JWT token for provided user.
func (s *UserService) authToken(u auth.User) (string, error) {
	claims := jwt.MapClaims{}
	claims["authorised"] = true
	claims["user_id"] = u.ID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)
	authToken, err := token.SignedString(tokenSecret)
	return authToken, err
}

// Create creates a new user.
// Returns an error if user with the same username or the same email
// already exist or if passwords do not match.
func (s *UserService) Create(u auth.UserRegister) error {
	if s.findByUsername(u.Username) != nil {
		return errors.New("user with such username already exists")
	}
	if s.findByEmail(u.Email) != nil {
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

	var usr auth.User
	rand.Seed(time.Now().UnixNano())
	usr.ID = rand.Uint64()
	usr.Username = u.Username
	usr.Email = strings.ToLower(u.Email)
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	usr.PasswordHash = string(hash)

	s.Users[usr.ID] = &usr
	return nil
}

// Authenticates a user by validating that it exists and hash of the
// provided password matches. On success returns a JWT token.
// While this method returns different errors for different failures the
// end user should only see a generic "invalid credentials" message.
func (s *UserService) Authenticate(u auth.UserLogin) (auth.UserInfo, error) {
	inf := auth.UserInfo{Username: "", Token: ""}
	usr := s.findByUsername(u.Username)
	if usr == nil {
		return inf, errors.New("user doesn't exist")
	}

	err := bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(u.Password))
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

// Validate checks if provided token is valid for the user.
func (s *UserService) Validate(u auth.User, t string) (bool, error) {
	token, err := jwt.Parse(t, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token signing method is not valid: %v", token.Header["alg"])
		}
		return tokenSecret, nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	} else {
		return false, fmt.Errorf("alg header %v, error: %v", claims["alg"], err)
	}
}
