package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	userMem "github.com/iliafrenkel/go-pb/src/api/auth/memory"
	"github.com/iliafrenkel/go-pb/src/api/base62"
	pasteMem "github.com/iliafrenkel/go-pb/src/api/paste/memory"
)

var pasteSvc api.PasteService = pasteMem.New()
var userSvc api.UserService = userMem.New()
var apiSrv *ApiServer
var mckSrv *httptest.Server

// createTestPaste creates a paste with a random ID and a random body.
func createTestPaste() *api.Paste {
	rand.Seed(time.Now().UnixNano())
	id := rand.Uint64()
	var p = api.Paste{
		ID:      id,
		Title:   "Test paste",
		Body:    base62.Encode(id),
		Expires: time.Time{},
	}

	return &p
}

func TestMain(m *testing.M) {
	apiSrv = New(pasteSvc, userSvc, ApiServerOptions{MaxBodySize: 10240})
	mckSrv = httptest.NewServer(apiSrv.Router)

	os.Exit(m.Run())
}

func Test_GetPaste(t *testing.T) {
	var paste = createTestPaste()
	if err := pasteSvc.Create(paste); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(mckSrv.URL + "/paste/" + paste.URL())

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status should be OK, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want, err := json.Marshal(paste)
	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_GetPasteNotFound(t *testing.T) {
	var paste = createTestPaste()

	resp, err := http.Get(mckSrv.URL + "/paste/" + paste.URL())

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status should be 404 Not Found, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := "paste not found"
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_GetPasteWrongID(t *testing.T) {
	resp, err := http.Get(mckSrv.URL + "/paste/SD)W*^W#4^&*S;!")

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status should be 404 Not Found, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := "paste not found"
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_CreatePaste(t *testing.T) {
	var paste = createTestPaste()

	want, _ := json.Marshal(paste)
	resp, err := http.Post(mckSrv.URL+"/paste", "application/json", bytes.NewBuffer(want))

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Status should be OK, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	body, _ := json.Marshal(paste.Body)
	if !strings.Contains(got, string(body)) {
		t.Errorf("Response should have body [%s], got [%s]", body, got)
	}
}

func Test_CreatePasteWrongContentType(t *testing.T) {
	var paste = createTestPaste()

	want, _ := json.Marshal(paste)
	resp, err := http.Post(mckSrv.URL+"/paste", "application/xml", bytes.NewBuffer(want))

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("Status should be %s, got %d", http.StatusText(http.StatusUnsupportedMediaType), resp.StatusCode)
	}
}

func Test_CreatePasteExtraField(t *testing.T) {
	var paste = createTestPaste()
	extraPaste := struct {
		api.Paste
		ExtraField string `json:"extraField"`
	}{
		*paste,
		"Extra field",
	}

	body, _ := json.Marshal(extraPaste)
	resp, err := http.Post(mckSrv.URL+"/paste", "application/json", bytes.NewBuffer(body))

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status should be %s, got %d", http.StatusText(http.StatusBadRequest), resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := fmt.Sprintf("request body contains unknown field \"%s\"", "extraField")
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_CreatePasteWrongJson(t *testing.T) {
	body := "this is not a json"
	resp, err := http.Post(mckSrv.URL+"/paste", "application/json", bytes.NewBuffer([]byte(body)))

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Status should be %s, got %d", http.StatusText(http.StatusBadRequest), resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := fmt.Sprintf("request body contains malformed JSON (at position %d)", 2)
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_DeletePaste(t *testing.T) {
	var paste = createTestPaste()
	if err := pasteSvc.Create(paste); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, mckSrv.URL+"/paste/"+paste.URL(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status should be OK, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := ""
	if got != string(want) {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_DeletePasteNotFound(t *testing.T) {
	var paste = createTestPaste()
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, mckSrv.URL+"/paste/"+paste.URL(), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Check status
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status should be %s, got %d", http.StatusText(http.StatusNotFound), resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := "paste not found"
	if got != string(want) {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_UserLogin(t *testing.T) {
	var ur = api.UserRegister{
		Username:   "test",
		Email:      "test@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	if err := userSvc.Create(ur); err != nil {
		t.Fatal(err)
	}

	// Login with correct username/password
	var ul = api.UserLogin{
		Username: ur.Username,
		Password: ur.Password,
	}
	data, _ := json.Marshal(ul)
	resp, err := http.Post(mckSrv.URL+"/user/login", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status should be OK, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var ui api.UserInfo
	json.Unmarshal(b, &ui)
	if ui.Username != ul.Username {
		t.Errorf("Response should have username [%s], got [%s]", ul.Username, ui.Username)
	}
	if ui.Token == "" {
		t.Errorf("Response should have token, got empty")
	}

	// Login with wrong password
	ul = api.UserLogin{
		Username: ur.Username,
		Password: "wrong",
	}
	data, _ = json.Marshal(ul)
	resp, err = http.Post(mckSrv.URL+"/user/login", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status should be Unauthorized, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	want := "Invalid credentials"
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}

	// Login with wrong username
	ul = api.UserLogin{
		Username: "wrong",
		Password: ur.Password,
	}
	data, _ = json.Marshal(ul)
	resp, err = http.Post(mckSrv.URL+"/user/login", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status should be Unauthorized, got %d", resp.StatusCode)
	}

	// Check body
	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	got = string(b)
	want = "Invalid credentials"
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_UserRegister(t *testing.T) {
	var ur = api.UserRegister{
		Username:   "test-register",
		Email:      "test-register@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	data, _ := json.Marshal(ur)
	resp, err := http.Post(mckSrv.URL+"/user/register", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status should be OK, got %d", resp.StatusCode)
	}

	// Register with existing username
	ur = api.UserRegister{
		Username:   "test-register",
		Email:      "test-register2@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	data, _ = json.Marshal(ur)
	resp, err = http.Post(mckSrv.URL+"/user/register", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Status should be Conflict, got %d", resp.StatusCode)
	}

	// Register with existing email
	ur = api.UserRegister{
		Username:   "test-register2",
		Email:      "test-register@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	data, _ = json.Marshal(ur)
	resp, err = http.Post(mckSrv.URL+"/user/register", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Status should be Conflict, got %d", resp.StatusCode)
	}
}

func Test_UserValidate(t *testing.T) {
	var ur = api.UserRegister{
		Username:   "test-validate",
		Email:      "test-validate@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	if err := userSvc.Create(ur); err != nil {
		t.Fatal(err)
	}
	var ul = api.UserLogin{
		Username: ur.Username,
		Password: ur.Password,
	}
	ui, err := userSvc.Authenticate(ul)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(mckSrv.URL+"/user/validate", "text/plain", bytes.NewBuffer([]byte(ui.Token)))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status should be OK, got %d", resp.StatusCode)
	}

	// Wrong token
	resp, err = http.Post(mckSrv.URL+"/user/validate", "text/plain", bytes.NewBuffer([]byte(ui.Token+"wrong")))
	if err != nil {
		t.Fatal(err)
	}
	// Check status
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Status should be Unauthorized, got %d", resp.StatusCode)
	}
}

// TODO:
//  - [ ] wrong HTTP methods for all endpoints
//  - [ ] multiple json objects for create
//  - [x] get paste with wrong id (not properly base62 encoded)
//  - [ ] test DeleteAfterRead
//  - [ ] test wrong JSON field type
//  - [ ] test create with empty body
