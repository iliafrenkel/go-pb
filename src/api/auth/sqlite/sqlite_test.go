package sqlite

import (
	"fmt"
	"os"
	"testing"

	"github.com/iliafrenkel/go-pb/src/api"
)

var usrSvc *UserService

func TestMain(m *testing.M) {
	var err error
	usrSvc, err = New(DBOptions{Connection: "file::memory:?cache=shared"})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func Test_CreateUser(t *testing.T) {
	var usr = api.UserRegister{
		Username:   "test",
		Email:      "test@example.com",
		Password:   "12345",
		RePassword: "12345",
	}

	err := usrSvc.Create(usr)

	if err != nil {
		t.Errorf("Failed to create a user: %v", err)
	}

	// Check if we can find the user by username
	u, err := usrSvc.findByUsername(usr.Username)
	if err != nil {
		t.Errorf("findByUsername failed: %w", err)
		return
	}
	if u == nil {
		t.Errorf("Failed to find a user by username")
	}
	// Check if we can find the user by email
	u, err = usrSvc.findByEmail(usr.Email)
	if err != nil {
		t.Errorf("findByUsername failed: %w", err)
		return
	}
	if u == nil {
		t.Errorf("Failed to find a user by email")
	}

	// Try to create with the same username but different email
	usr.Email = "another@example.com"
	err = usrSvc.Create(usr)
	if err == nil {
		t.Errorf("Succeeded to create a user with existing username")
	}
	// Try to create with the same email but different username
	usr.Email = "test@example.com"
	usr.Username = "test2"
	err = usrSvc.Create(usr)
	if err == nil {
		t.Errorf("Succeeded to create a user with existing email")
	}
}

func Test_CreateUserEmptyUsername(t *testing.T) {
	var usr = api.UserRegister{
		Username:   "",
		Email:      "emptyusername@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	err := usrSvc.Create(usr)
	if err == nil {
		t.Errorf("Succeeded to create a user with empty username")
	}
}
func Test_CreateUserEmptyEmail(t *testing.T) {
	var usr = api.UserRegister{
		Username:   "emptyemail",
		Email:      "",
		Password:   "12345",
		RePassword: "12345",
	}
	err := usrSvc.Create(usr)
	if err == nil {
		t.Errorf("Succeeded to create a user with empty email")
	}
}

func Test_CreateUserPasswordsDontMatch(t *testing.T) {
	var usr = api.UserRegister{
		Username:   "nonmatchingpasswords",
		Email:      "nonmatchingpasswords@example.com",
		Password:   "12345",
		RePassword: "54321",
	}
	err := usrSvc.Create(usr)
	if err == nil {
		t.Errorf("Succeeded to create a user with non-matching passwords")
	}
}

func Test_AuthenticateUser(t *testing.T) {
	var usr = api.UserRegister{
		Username:   "auth",
		Email:      "auth@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	err := usrSvc.Create(usr)
	if err != nil {
		t.Errorf("Failed to create a user: %v", err)
	}

	var usrLogin = api.UserLogin{
		Username: usr.Username,
		Password: usr.Password,
	}

	inf, err := usrSvc.Authenticate(usrLogin)

	if err != nil {
		t.Errorf("Failed to authenticate a user: %v", err)
	}

	if err == nil && inf.Token == "" {
		t.Errorf("Failed to authenticate a user: error is nil but token is empty")
	}

	//user doesn't exist
	usrLogin = api.UserLogin{
		Username: "idontexist",
		Password: "idontmatter",
	}

	_, err = usrSvc.Authenticate(usrLogin)

	if err == nil {
		t.Errorf("Authentication succeeded for a user that doesn't exist: %#v", usrLogin)
	}

	//wrong password
	usrLogin = api.UserLogin{
		Username: usr.Username,
		Password: "wrong",
	}
	_, err = usrSvc.Authenticate(usrLogin)

	if err == nil {
		t.Errorf("Authentication succeeded with incorrect password: %#v", usrLogin)
	}
}

func Test_ValidateToken(t *testing.T) {
	var usr = api.UserRegister{
		Username:   "validate",
		Email:      "validate@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	err := usrSvc.Create(usr)
	if err != nil {
		t.Errorf("Failed to create a user: %v", err)
	}

	var usrLogin = api.UserLogin{
		Username: usr.Username,
		Password: usr.Password,
	}

	inf, err := usrSvc.Authenticate(usrLogin)

	if err != nil {
		t.Errorf("Failed to authenticate a user: %v", err)
	}

	u, err := usrSvc.findByUsername(usr.Username)
	if err != nil {
		t.Errorf("findByUsername failed: %w", err)
		return
	}
	v, err := usrSvc.Validate(*u, inf.Token)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}
	if v.Username == "" || v.Token == "" {
		t.Errorf("Token validation failed: %s - %#v", inf.Token, v)

	}
}
