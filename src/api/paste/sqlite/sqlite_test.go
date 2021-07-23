package sqlite

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// createTestPaste create a paste with random ID and Body
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

func Test_Create(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	var p = createTestPaste()
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Title != p.Title {
			t.Errorf("wrong title, want %s got %s", p.Title, paste.Title)
		}
		if paste.Body != p.Body {
			t.Errorf("wrong body, want %s got %s", p.Title, paste.Title)
		}
	}
}

func Test_CreateWithExpiration(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	var p = createTestPaste()
	// Minutes
	p.Expires = "10m"
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(10*time.Minute) {
			t.Errorf("minutes, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
	// Hours
	p.Expires = "2h"
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(2*time.Hour) {
			t.Errorf("hours, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
	// Days
	p.Expires = "2d"
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(48*time.Hour) {
			t.Errorf("days, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
	// Weeks
	p.Expires = "1w"
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(7*24*time.Hour) {
			t.Errorf("weeks, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
	// Months
	p.Expires = "6M"
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		y1, m1, _ := paste.Expires.Date()
		y2, m2, _ := paste.Created.Date()
		yearDiff := (y1 - y2) * 12
		if int(m1-m2)+yearDiff != 6 {
			t.Errorf("months, wrong expiration: created %s[%d], expires %s[%d]", paste.Created, m2, paste.Expires, m1)
		}
	}
	// Years
	p.Expires = "1y"
	if paste, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		y1, _, _ := paste.Expires.Date()
		y2, _, _ := paste.Created.Date()
		if y1-y2 != 1 {
			t.Errorf("years, wrong expiration: created %s[%d], expires %s[%d]", paste.Created, y2, paste.Expires, y1)
		}
	}
	// Unknown format
	p.Expires = "1z"
	if _, err := pasteSvc.Create(*p); err == nil {
		t.Error("paste created successfully but shouldn't")
	} else {
		got := err.Error()
		want := "unknown duration format: 1z"
		if want != got {
			t.Errorf("expected %s, got %s", want, got)
		}
	}
	// Wrong format
	p.Expires = "abc"
	if _, err := pasteSvc.Create(*p); err == nil {
		t.Error("paste created successfully but shouldn't")
	} else {
		got := err.Error()
		want := "wrong duration format: abc"
		if !strings.HasPrefix(got, want) {
			t.Errorf("expected %s to start with %s", got, want)
		}
	}
}

func Test_GetPaste(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	var p = createTestPaste()
	paste, err := pasteSvc.Create(*p)
	if err != nil {
		t.Errorf("failed to create a paste: %v", err)
	}

	_, err = pasteSvc.Get(paste.ID)
	if err != nil {
		t.Errorf("failed to find a paste: %v", err)
	}
}

func Test_PasteNotFound(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	p, err := pasteSvc.Get(0)
	if err != nil {
		t.Errorf("failed to get a paste: %v", err)
	} else {
		if p != nil {
			t.Errorf("expect paste to not exist, got %#v", p)
		}
	}
}

func Test_Delete(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	var p = createTestPaste()
	paste, err := pasteSvc.Create(*p)
	if err != nil {
		t.Errorf("failed to create a paste: %v", err)
	}

	err = pasteSvc.Delete(paste.ID)
	if err != nil {
		t.Errorf("Failed to delete a paste: %v", err)
	}

	paste, err = pasteSvc.Get(paste.ID)
	if err != nil {
		t.Errorf("failed to get a paste: %v", err)
	}
	if paste != nil {
		t.Errorf("Found a paste after delete: %#v", paste)
	}
}

func Test_DeleteNotFound(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	err = pasteSvc.Delete(0)
	if err != nil {
		t.Errorf("failed to delete a non-existsing paste: %v", err)
	}
}

func Test_List(t *testing.T) {
	t.Parallel()

	var err error
	var pasteSvc *PasteService
	var db *gorm.DB
	db, err = gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		fmt.Printf("Failed to create a UserService: %v\n", err)
		os.Exit(1)
	}
	pasteSvc, err = New(SvcOptions{DBConnection: db})
	if err != nil {
		t.Fatalf("Failed to create a PasteService: %v\n", err)
	}

	var p = createTestPaste()
	p.UserID = 1
	if _, err := pasteSvc.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
		return
	}

	list := pasteSvc.List(1)
	if len(list) != 1 {
		t.Errorf("Expected a list of 1, got %d/n%v", len(list), list)
		return
	}
	if p.Title != list[0].Title {
		t.Errorf("wrong title, want %s got %s", p.Title, list[0].Title)
	}
}
