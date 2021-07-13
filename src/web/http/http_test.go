package http

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/base62"
	"github.com/iliafrenkel/go-pb/src/api/db/memory"
	apihttp "github.com/iliafrenkel/go-pb/src/api/http"
)

var webSrv *WebServer
var apiSrv *apihttp.ApiServer
var memSvc api.PasteService = memory.New()
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
	os.Chdir("../../") // Needed for proper template loading
	apiSrv = apihttp.New(memSvc)
	mckSrv = httptest.NewServer(apiSrv.Router)
	webSrv = New(WebServerOptions{ApiURL: mckSrv.URL})

	os.Exit(m.Run())
}

func Test_RootRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be OK, got %d", w.Code)
	}

	want := "Go PB - Home"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_PingRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/ping", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be OK, got %d", w.Code)
	}

	want := "{\"message\":\"pong\"}"
	got := w.Body.String()
	if got != want {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_UserLoginRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/u/login", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be OK, got %d", w.Code)
	}

	want := "Go PB - Login"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_UserRegisterRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/u/register", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be OK, got %d", w.Code)
	}

	want := "Go PB - Register"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_PasteNotFoundRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/fakeid", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status should be %d, got %d", http.StatusNotFound, w.Code)
	}

	want := "Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_PasteRoute(t *testing.T) {
	p := createTestPaste()
	if err := memSvc.Create(p); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/"+p.URL(), nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := "Test paste"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_PasteCreateRoute(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test create paste")
	form.Set("body", "This is a test create paste")
	form.Set("syntax", "none")
	form.Set("delete_after_read", "false")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := "Test create paste"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_PasteCreateNoBodyRoute(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test create paste")
	form.Set("syntax", "none")
	form.Set("delete_after_read", "false")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status should be %d, got %d", http.StatusBadRequest, w.Code)
	}

	want := "Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_PasteCreateNoSyntaxRoute(t *testing.T) {
	form := url.Values{}
	form.Set("title", "Test create paste")
	form.Set("Body", "This is a test create paste")
	form.Set("delete_after_read", "false")
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status should be %d, got %d", http.StatusBadRequest, w.Code)
	}

	want := "Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}
