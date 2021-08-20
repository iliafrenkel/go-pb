package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/lgr"
	"github.com/iliafrenkel/go-pb/src/service"
	"github.com/iliafrenkel/go-pb/src/store"
)

var webSrv *WebServer

// TestMain is a setup function for the test suite. It creates a new WebServer
// with options suitable for testing.
func TestMain(m *testing.M) {
	log := lgr.New(lgr.Debug, lgr.CallerFile, lgr.CallerFunc, lgr.Msec, lgr.LevelBraces)

	webSrv = New(log, WebServerOptions{
		Addr:               "localhost:8080",
		Proto:              "http",
		ReadTimeout:        2,
		WriteTimeout:       2,
		IdleTimeout:        5,
		LogFile:            "",
		LogMode:            "debug",
		MaxBodySize:        1024,
		BrandName:          "Go PB",
		BrandTagline:       "Testing is good!",
		Assets:             "../../assets",
		Templates:          "../../templates",
		Version:            "test",
		AuthSecret:         "ki7GZphH7bRNhKN8476jUTJn2QaMRxhX",
		AuthTokenDuration:  60 * time.Second,
		AuthCookieDuration: 60 * time.Second,
		AuthIssuer:         "go-pb test",
		AuthURL:            "http://localhost:8080",
		DBType:             "memory",
	})

	os.Exit(m.Run())
}

// TestGetHomePage verifies the GET / route handler. It checks that the home
// page is generated with correct title and that the New Paste form is there.
func TestGetHomePage(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := webSrv.options.BrandName + " - Home"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = `<form method="POST" action="/p/">`
	if !strings.Contains(got, want) {
		t.Errorf("Response should have form [%s], got [%s]", want, got)
	}
}

// TestPostPasteDefaults create a paste with just the required fields.
func TestPostPasteDefaults(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	form := url.Values{}
	form.Add("body", "Test body")
	form.Add("privacy", "public")
	req, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	webSrv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := webSrv.options.BrandName + " - Paste"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = `<code class="py-3 language-text">Test body</code>`
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// TestPostPasteEmptyForm try to POST an empty form
func TestPostPasteEmptyForm(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	form := url.Values{}
	req, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	webSrv.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status should be %d, got %d", http.StatusBadRequest, w.Code)
	}

	want := webSrv.options.BrandName + " - Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = "Body must not be empty"
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Wrong value for privacy
func TestPostPasteWrongPrivacy(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	form := url.Values{}
	form.Add("body", "Test body")
	form.Add("privacy", "absolutely public")
	req, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	webSrv.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status should be %d, got %d", http.StatusBadRequest, w.Code)
	}

	want := webSrv.options.BrandName + " - Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = "Privacy can be one of &#39;private&#39;, &#39;public&#39; or &#39;unlisted&#39;."
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Wrong value for expiration
func TestPostPasteWrongExpiration(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	form := url.Values{}
	form.Add("body", "Test body")
	form.Add("privacy", "public")
	form.Add("expires", "1,3z")
	req, _ := http.NewRequest("POST", "/p/", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	webSrv.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Status should be %d, got %d", http.StatusBadRequest, w.Code)
	}

	want := webSrv.options.BrandName + " - Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = "Duration format is incorrect."
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// TestNotFoundPage verifies the NotFound handler. It checks that the error
// page has the correct title and error message and that there is a link to
// the home page.
func TestNotFoundPage(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/NotFoundPage", nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status should be %d, got %d", http.StatusNotFound, w.Code)
	}

	want := webSrv.options.BrandName + " - Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = "Unfortunately the page you are looking for is not there 🙁"
	if !strings.Contains(got, want) {
		t.Errorf("Response should have error message [%s], got [%s]", want, got)
	}

	want = `<a href="/"`
	if !strings.Contains(got, want) {
		t.Errorf("Response should have link to home [%s], got [%s]", want, got)
	}
}

// Get public paste
func TestGetPublicPaste(t *testing.T) {
	t.Parallel()

	p, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test",
		Body:            "Test paste",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "",
		Syntax:          "text",
		UserID:          "",
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/"+p.URL(), nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := webSrv.options.BrandName + " - Paste"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = `<code class="py-3 language-text">Test paste</code>`
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Get non-existing paste
func TestGetNonExistingPaste(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/IYCE8rJj8Qg", nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status should be %d, got %d", http.StatusNotFound, w.Code)
	}

	want := webSrv.options.BrandName + " - Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = "There is no such paste"
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Get private paste of another user
func TestGetPrivatePasteOfAnotherUser(t *testing.T) {
	t.Parallel()

	u, _ := webSrv.service.GetOrUpdateUser(store.User{
		ID:   "test_user",
		Name: "Test User",
	})
	p, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test",
		Body:            "Test paste",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "private",
		Password:        "",
		Syntax:          "text",
		UserID:          u.ID,
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/"+p.URL(), nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status should be %d, got %d", http.StatusForbidden, w.Code)
	}

	want := webSrv.options.BrandName + " - Error"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = "This paste is private"
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Get private paste of the user who created it
func TestGetPrivatePaste(t *testing.T) {
	t.Parallel()

	u, _ := webSrv.service.GetOrUpdateUser(store.User{
		ID:   "test_user_1",
		Name: "Test User 1",
	})
	p, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test",
		Body:            "Test paste",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "private",
		Password:        "",
		Syntax:          "text",
		UserID:          u.ID,
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/"+p.URL(), nil)
	// Add user to request context
	r = token.SetUserInfo(r, token.User{
		Name: u.Name,
		ID:   u.ID,
	})
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := webSrv.options.BrandName + " - Paste"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = `<code class="py-3 language-text">Test paste</code>`
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Get password protected paste without password
func TestGetPasswordProtectedPasteNoPassword(t *testing.T) {
	t.Parallel()

	p, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test",
		Body:            "Test paste",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "pa$$w0rd",
		Syntax:          "text",
		UserID:          "",
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/"+p.URL(), nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status should be %d, got %d", http.StatusUnauthorized, w.Code)
	}

	want := webSrv.options.BrandName + " - Password"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = `<form method="POST" action="/p/` + p.URL()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Get password protected paste with password
func TestGetPasswordProtectedPasteWithPassword(t *testing.T) {
	t.Parallel()

	p, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test",
		Body:            "Test paste",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "pa$$w0rd",
		Syntax:          "text",
		UserID:          "",
	})

	w := httptest.NewRecorder()
	form := url.Values{}
	form.Add("password", "pa$$w0rd")
	r, _ := http.NewRequest("POST", "/p/"+p.URL(), strings.NewReader(form.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := webSrv.options.BrandName + " - Paste"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = `<code class="py-3 language-text">Test paste</code>`
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

// Get a list of pastes for a user
func TestGetUserPaste(t *testing.T) {
	t.Parallel()

	p1, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test 1",
		Body:            "Test paste 1",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "",
		Syntax:          "text",
		UserID:          "",
	})
	p2, _ := webSrv.service.NewPaste(service.PasteRequest{
		Title:           "Test 2",
		Body:            "Test paste 2",
		Expires:         "",
		DeleteAfterRead: false,
		Privacy:         "private",
		Password:        "",
		Syntax:          "text",
		UserID:          "",
	})

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/l/", nil)
	webSrv.router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := webSrv.options.BrandName + " - Pastes"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have title [%s], got [%s]", want, got)
	}

	want = p1.Title
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}

	want = p2.Title
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}
