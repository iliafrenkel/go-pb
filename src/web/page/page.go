package page

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"github.com/go-pkgz/auth/token"
	"github.com/iliafrenkel/go-pb/src/store"
)

// Page type represent a single HTML page with all the data that any page
// might need. Most pages won't need all of the data options. Use Data
// functions defined below to add data to the page.
type Page struct {
	// common for all pages
	Title   string // page title, used a value for the <title> tag
	Brand   string // text displayed in big letters at the top of each page
	Tagline string // text displayed below the Brand
	Logo    string // name of the image file from the assets folder to use as a logo
	Theme   string // bootstrap theme
	Server  string // server URL
	Version string // application version to show at the bottom of every page
	Totals  Stats  // totals, such as total number of pastes and users

	// not common for all pages
	User      token.User    // user details parsed from the JWT token
	PasteID   string        // paste ID (URL) for pages that need redirect/post back
	Pastes    []store.Paste // a list of pastes
	Paste     store.Paste   // a single paste
	PageLinks Paginator     // paginator for list pages
	LastPage  int           // offset for the last paginator link

	// only for error pages
	ErrorCode    int    // error code, to show on the error page (404, 500, etc.)
	ErrorText    string // error text, friendly text to accompany the error code
	ErrorMessage string // optional error message to help the user with what to do next

	// for internal use
	templates *template.Template // all the loaded templates from the server
	template  string             // template name to generate HTML
}

// Data func type.
type Data func(p *Page)

// Stats application-wide statistics.
type Stats struct {
	Pastes int64
	Users  int64
}

// PaginatorLink contains all the data needed to construct a single paginator link.
type PaginatorLink struct {
	Number int // page number
	Offset int // offset for the page
}

// Paginator struct used to build paginators on list pages.
type Paginator struct {
	Current    int             // current page number
	Last       int             // last page number
	LastOffset int             // last page number
	Pages      []PaginatorLink // a list of links
}

// Title sets page title.
func Title(title string) Data {
	return func(p *Page) {
		p.Title = title
	}
}

// Brand sets page brand.
func Brand(brand string) Data {
	return func(p *Page) {
		p.Brand = brand
	}
}

// Tagline sets page brand.
func Tagline(tagline string) Data {
	return func(p *Page) {
		p.Tagline = tagline
	}
}

// Logo sets page logo.
func Logo(logo string) Data {
	return func(p *Page) {
		p.Logo = logo
	}
}

// Theme sets page theme.
func Theme(theme string) Data {
	return func(p *Page) {
		p.Theme = theme
	}
}

// Server sets page server.
func Server(server string) Data {
	return func(p *Page) {
		p.Server = server
	}
}

// Version sets page version.
func Version(version string) Data {
	return func(p *Page) {
		p.Version = version
	}
}

// Totals sets page totals.
func Totals(totals Stats) Data {
	return func(p *Page) {
		p.Totals = totals
	}
}

// User sets page user.
func User(usr token.User) Data {
	return func(p *Page) {
		p.User = usr
	}
}

// PasteID sets paste ID for pages that need it.
func PasteID(pid string) Data {
	return func(p *Page) {
		p.PasteID = pid
	}
}

// Pastes sets a list of pastes.
func Pastes(pastes []store.Paste) Data {
	return func(p *Page) {
		p.Pastes = pastes
	}
}

// Paste sets a single paste.
func Paste(paste store.Paste) Data {
	return func(p *Page) {
		p.Paste = paste
	}
}

// PageLinks sets paginator for the page.
func PageLinks(paginator Paginator) Data {
	return func(p *Page) {
		p.PageLinks = paginator
	}
}

// ErrorCode sets error code for the error page.
func ErrorCode(code int) Data {
	return func(p *Page) {
		p.ErrorCode = code
	}
}

// ErrorText sets error text for the error page.
func ErrorText(txt string) Data {
	return func(p *Page) {
		p.ErrorText = txt
	}
}

// ErrorMessage sets error message for the error page.
func ErrorMessage(msg string) Data {
	return func(p *Page) {
		p.ErrorMessage = msg
	}
}

// Template sets the template name for the page.
func Template(name string) Data {
	return func(p *Page) {
		p.template = name
	}
}

// New returns a new page.
func New(t *template.Template, data ...Data) *Page {
	p := Page{
		templates: t,
	}
	for _, d := range data {
		d(&p)
	}

	return &p
}

// Show renders the template with the page data and writes resulting HTML.
func (p *Page) Show(w io.Writer) error {
	var html bytes.Buffer
	err := p.templates.ExecuteTemplate(&html, p.template, p)
	if err != nil {
		return fmt.Errorf("ERROR error executing template: %w", err)
	}

	_, err = w.Write(html.Bytes())
	return err
}
