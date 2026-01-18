package page

import (
	"bytes"
	"html/template"
	"reflect"
	"testing"

	"github.com/go-pkgz/auth/v2/token"
	"github.com/iliafrenkel/go-pb/src/store"
)

func TestNew(t *testing.T) {
	p := New(nil, Brand("Test"))

	if p.Brand != "Test" {
		t.Errorf("expected brand to be 'Test', got '%s'", p.Brand)
	}
}

func TestPageData(t *testing.T) {
	user := token.User{
		Name: "test",
	}
	paste := store.Paste{
		Title: "test paste",
	}
	pastes := []store.Paste{paste}
	paginator := Paginator{
		Current: 1,
	}
	stats := Stats{
		Pastes: 1,
		Users:  1,
	}
	tests := []struct {
		name string
		data Data
		want Page
	}{
		{"Title", Title("test title"), Page{Title: "test title"}},
		{"Brand", Brand("test brand"), Page{Brand: "test brand"}},
		{"Tagline", Tagline("test tagline"), Page{Tagline: "test tagline"}},
		{"Logo", Logo("test logo"), Page{Logo: "test logo"}},
		{"Theme", Theme("test theme"), Page{Theme: "test theme"}},
		{"Server", Server("test server"), Page{Server: "test server"}},
		{"Version", Version("test version"), Page{Version: "test version"}},
		{"Totals", Totals(stats), Page{Totals: stats}},
		{"User", User(user), Page{User: user}},
		{"PasteID", PasteID("test_paste_id"), Page{PasteID: "test_paste_id"}},
		{"Pastes", Pastes(pastes), Page{Pastes: pastes}},
		{"UserPastes", UserPastes(pastes), Page{UserPastes: pastes}},
		{"Paste", Paste(paste), Page{Paste: paste}},
		{"PageLinks", PageLinks(paginator), Page{PageLinks: paginator}},
		{"ErrorCode", ErrorCode(404), Page{ErrorCode: 404}},
		{"ErrorText", ErrorText("test error text"), Page{ErrorText: "test error text"}},
		{"ErrorMessage", ErrorMessage("test error message"), Page{ErrorMessage: "test error message"}},
		{"Template", Template("test_template"), Page{template: "test_template"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Page{}
			tt.data(p)
			if !reflect.DeepEqual(*p, tt.want) {
				t.Errorf("got = %v, want %v", *p, tt.want)
			}
		})
	}
}

func TestPage_Show(t *testing.T) {
	tmpl, err := template.New("test").Parse(`{{.Title}}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	p := New(tmpl, Title("test page"), Template("test"))

	var buf bytes.Buffer
	err = p.Show(&buf)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buf.String() != "test page" {
		t.Errorf("expected 'test page', got '%s'", buf.String())
	}
}

func TestPage_ShowError(t *testing.T) {
	tmpl, err := template.New("test").Parse(`{{.Title}}`)
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}
	// give a wrong template name to trigger an error
	p := New(tmpl, Title("test page"), Template("wrong"))

	var buf bytes.Buffer
	err = p.Show(&buf)
	if err == nil {
		t.Error("expected an error, got none")
	}
}
