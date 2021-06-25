package http

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/api"
	"github.com/iliafrenkel/go-pb/api/memory"
)

var svc api.PasteService = memory.New()

// createTestPaste creates a paste with a random ID and a random body.
func createTestPaste() *api.Paste {
	b := make([]byte, 16)
	rand.Read(b)
	var p = api.Paste{
		ID:      fmt.Sprintf("%x", md5.Sum(b)),
		Title:   "Test paste",
		Body:    b,
		Expires: time.Time{},
	}

	return &p
}

func Test_GetPaste(t *testing.T) {
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()
	if err := svc.Create(paste); err != nil {
		t.Fatal(err)
	}

	// Documentation : https://golang.org/pkg/net/http/httptest/#NewServer
	mockServer := httptest.NewServer(apiSrv.Router)
	resp, err := http.Get(mockServer.URL + "/paste/" + paste.ID)

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
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()

	mockServer := httptest.NewServer(apiSrv.Router)
	resp, err := http.Get(mockServer.URL + "/paste/" + paste.ID)

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
	want := "paste not found\n"
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_CreatePaste(t *testing.T) {
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()

	mockServer := httptest.NewServer(apiSrv.Router)
	want, _ := json.Marshal(paste)
	resp, err := http.Post(mockServer.URL+"/paste", "application/json", bytes.NewBuffer(want))

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
	body, _ := json.Marshal(paste.Body)
	if !strings.Contains(got, string(body)) {
		t.Errorf("Response should have body [%s], got [%s]", body, got)
	}
}

func Test_CreatePasteWrongContentType(t *testing.T) {
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()

	mockServer := httptest.NewServer(apiSrv.Router)
	want, _ := json.Marshal(paste)
	resp, err := http.Post(mockServer.URL+"/paste", "application/xml", bytes.NewBuffer(want))

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
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()
	extraPaste := struct {
		api.Paste
		ExtraField string `json:"extraField"`
	}{
		*paste,
		"Extra field",
	}

	mockServer := httptest.NewServer(apiSrv.Router)
	body, _ := json.Marshal(extraPaste)
	resp, err := http.Post(mockServer.URL+"/paste", "application/json", bytes.NewBuffer(body))

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
	want := fmt.Sprintf("Request body contains unknown field \"%s\"\n", "extraField")
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_CreatePasteWrongJson(t *testing.T) {
	var apiSrv *ApiServer = New(svc)

	mockServer := httptest.NewServer(apiSrv.Router)
	body := "this is not a json"
	resp, err := http.Post(mockServer.URL+"/paste", "application/json", bytes.NewBuffer([]byte(body)))

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
	want := fmt.Sprintf("Request body contains malformed JSON (at position %d)\n", 2)
	if got != want {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

func Test_DeletePaste(t *testing.T) {
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()
	if err := svc.Create(paste); err != nil {
		t.Fatal(err)
	}
	mockServer := httptest.NewServer(apiSrv.Router)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, mockServer.URL+"/paste/"+paste.ID, nil)
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
	var apiSrv *ApiServer = New(svc)
	var paste = createTestPaste()
	mockServer := httptest.NewServer(apiSrv.Router)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, mockServer.URL+"/paste/"+paste.ID, nil)
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
	want := "paste not found\n"
	if got != string(want) {
		t.Errorf("Response should be [%s], got [%s]", want, got)
	}
}

// TODO:
//  - wrong HTTP methods for all endpoints
//  - multiple json objects for create
