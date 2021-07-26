package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/iliafrenkel/go-pb/src/api"
	userMem "github.com/iliafrenkel/go-pb/src/api/auth/memory"
	apihttp "github.com/iliafrenkel/go-pb/src/api/http"
	pasteMem "github.com/iliafrenkel/go-pb/src/api/paste/memory"
)

var webSrv *WebServer
var apiSrv *apihttp.APIServer
var pasteSvc api.PasteService = pasteMem.New()
var userSvc api.UserService = userMem.New()
var mckSrv *httptest.Server

// createTestPaste creates a paste with a random ID and a random body.
func createTestPaste() *api.PasteForm {
	var p = api.PasteForm{
		Title:           "Test paste",
		Body:            "Test body",
		Expires:         "never",
		DeleteAfterRead: false,
		Password:        "",
		Syntax:          "none",
		UserID:          0,
	}

	return &p
}

func TestMain(m *testing.M) {
	os.Chdir("../../") // Needed for proper template loading
	apiSrv = apihttp.New(pasteSvc, userSvc, apihttp.APIServerOptions{MaxBodySize: 10240})
	mckSrv = httptest.NewServer(apiSrv.Router)
	webSrv = New(WebServerOptions{
		Addr:          "localhost:8080",
		Proto:         "http",
		APIURL:        mckSrv.URL,
		ReadTimeout:   15,
		WriteTimeout:  15,
		IdleTimeout:   60,
		LogFile:       "",
		LogMode:       "debug",
		CookieAuthKey: "test",
		BrandName:     "Go PB",
		BrandTagline:  "A nice and simple pastebin alternative that you can host yourself.",
		Assets:        "./web/assets",
		Templates:     "./web/templates",
		Logo:          "bighead.svg",
		Version:       "v0.0.0-test",
	})

	os.Exit(m.Run())
}

func Test_RootRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be OK, got %d", w.Code)
	}

	want := webSrv.Options.BrandName + " - Home"
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_NoRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/noroute", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Status should be %d, got %d", http.StatusNotFound, w.Code)
	}

	want := webSrv.Options.BrandName + " - Error"
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

	want := webSrv.Options.BrandName + " - Login"
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

	want := webSrv.Options.BrandName + " - Register"
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
	paste, err := pasteSvc.Create(*p)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/"+paste.URL(), nil)
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
	form.Set("expires", "never")
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
	form.Set("expires", "never")
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
	form.Set("body", "This is a test create paste")
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

func Test_PasteList(t *testing.T) {
	p := createTestPaste()
	_, err := pasteSvc.Create(*p)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/p/list", nil)
	webSrv.Router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Status should be %d, got %d", http.StatusOK, w.Code)
	}

	want := "You will see all your pastes here once you login" //paste.Title
	got := w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_DoUserLoginRoute(t *testing.T) {
	// Create a test user
	var ur = api.UserRegister{
		Username:   "test",
		Email:      "test@example.com",
		Password:   "12345",
		RePassword: "12345",
	}
	if err := userSvc.Create(ur); err != nil {
		t.Fatal(err)
	}
	// "Fill" the login form with the correct details
	form := url.Values{}
	form.Set("username", "test")
	form.Set("password", "12345")
	// Simulate the POST request to the login page and record the response
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/u/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)
	cookie := w.Result().Header.Get("Set-Cookie")
	// Check the response status code
	if w.Code != http.StatusFound {
		t.Errorf("Status should be %d, got %d", http.StatusFound, w.Code)
	}
	// Check that we were redirected to the homepage
	want := "/"
	got := w.Result().Header.Get("Location")
	if !strings.Contains(got, want) {
		t.Errorf("The Location header should be [%s], got [%s]", want, got)
	}
	// Check if we are logged in
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("GET", "/", nil)
	r.Header.Set("Cookie", "token="+cookie)
	webSrv.Router.ServeHTTP(w, r)
	want = "test"
	got = w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should contain the username [%s], got [%s]", want, got)
	}

	// Try to login with a wrong password
	// "Fill" the login form with the correct details
	form.Set("username", "test")
	form.Set("password", "54321")
	// Simulate the POST request to the login page and record the response
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/u/login", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)
	// Check the response status code
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status should be %d, got %d", http.StatusUnauthorized, w.Code)
	}
	// Check that the login page has generic error message
	want = "Either username or password is incorrect"
	got = w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_DoUserRegisterRoute(t *testing.T) {
	// "Fill" the registration form with the correct details
	form := url.Values{}
	form.Set("username", "test-register")
	form.Set("email", "test-register@example.com")
	form.Set("password", "12345")
	form.Set("repassword", "12345")
	// Simulate the POST request to the register page and record the response
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/u/register", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)
	// Check the response status code
	if w.Code != http.StatusFound {
		t.Errorf("Status should be %d, got %d", http.StatusFound, w.Code)
	}
	// Check that we were redirected to the login page
	want := "/u/login"
	got := w.Result().Header.Get("Location")
	if !strings.Contains(got, want) {
		t.Errorf("The Location header should be [%s], got [%s]", want, got)
	}

	// Try to register the same user
	// Simulate the POST request to the login page and record the response
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/u/register", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)
	// Check the response status code
	if w.Code != http.StatusConflict {
		t.Errorf("Status should be %d, got %d", http.StatusConflict, w.Code)
	}
	// Check that the register page has the correct error message
	want = "User with such username already exists"
	got = w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}

	// Try to register a user when passwords do not match
	form.Set("username", "test-wrong-password")
	form.Set("email", "test-wrong-password@example.com")
	form.Set("password", "12345")
	form.Set("repassword", "wrong")
	// Simulate the POST request to the login page and record the response
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/u/register", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	webSrv.Router.ServeHTTP(w, r)
	// Check the response status code
	if w.Code != http.StatusConflict {
		t.Errorf("Status should be %d, got %d", http.StatusConflict, w.Code)
	}
	// Check that the register page has the correct error message
	want = "Passwords don&#39;t match"
	got = w.Body.String()
	if !strings.Contains(got, want) {
		t.Errorf("Response should have body [%s], got [%s]", want, got)
	}
}

func Test_UserLogoutRoute(t *testing.T) {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/u/logout", nil)
	webSrv.Router.ServeHTTP(w, r)
	// Check the status
	if w.Code != http.StatusFound {
		t.Errorf("Status should be %d, got %d", http.StatusFound, w.Code)
	}
	// Check that we were redirected to the homepage
	want := "/"
	got := w.Result().Header.Get("Location")
	if !strings.Contains(got, want) {
		t.Errorf("The Location header should be [%s], got [%s]", want, got)
	}
}
