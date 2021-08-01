package memory

import (
	"strings"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
)

var service = New()

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
	var p = createTestPaste()
	if paste, err := service.Create(*p); err != nil {
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

func Test_CreateWithExpirationMinutes(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Minutes
	p.Expires = "10m"
	if paste, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(10*time.Minute) {
			t.Errorf("minutes, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
}
func Test_CreateWithExpirationHours(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Hours
	p.Expires = "2h"
	if paste, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(2*time.Hour) {
			t.Errorf("hours, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
}
func Test_CreateWithExpirationDays(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Days
	p.Expires = "2d"
	if paste, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(48*time.Hour) {
			t.Errorf("days, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
}
func Test_CreateWithExpirationWeeks(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Weeks
	p.Expires = "1w"
	if paste, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		if paste.Expires.Sub(paste.Created) != time.Duration(7*24*time.Hour) {
			t.Errorf("weeks, wrong expiration: created %s, expires %s", paste.Created, paste.Expires)
		}
	}
}
func Test_CreateWithExpirationMonths(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Months
	p.Expires = "6M"
	if paste, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		y1, m1, _ := paste.Expires.Date()
		y2, m2, _ := paste.Created.Date()
		yearDiff := (y1 - y2) * 12
		if int(m1-m2)+yearDiff != 6 {
			t.Errorf("months, wrong expiration: created %s[%d], expires %s[%d]", paste.Created, m2, paste.Expires, m1)
		}
	}
}
func Test_CreateWithExpirationYears(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Years
	p.Expires = "1y"
	if paste, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
	} else {
		y1, _, _ := paste.Expires.Date()
		y2, _, _ := paste.Created.Date()
		if y1-y2 != 1 {
			t.Errorf("years, wrong expiration: created %s[%d], expires %s[%d]", paste.Created, y2, paste.Expires, y1)
		}
	}
}
func Test_CreateWithExpirationWrongFormat(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	// Unknown format
	p.Expires = "1z"
	if _, err := service.Create(*p); err == nil {
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
	if _, err := service.Create(*p); err == nil {
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
	var p = createTestPaste()
	paste, err := service.Create(*p)
	if err != nil {
		t.Errorf("failed to create a paste: %v", err)
	}

	_, err = service.Get(paste.ID)
	if err != nil {
		t.Errorf("failed to find a paste: %v", err)
	}
}

func Test_PasteNotFound(t *testing.T) {
	t.Parallel()
	p, err := service.Get(0)
	if err != nil {
		t.Errorf("Unexpected error, got %v", err)
		return
	}
	if p != nil {
		t.Errorf("Expected paste to be nil, got %#v", p)
	}
}

func Test_Delete(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	paste, err := service.Create(*p)
	if err != nil {
		t.Errorf("failed to create a paste: %v", err)
	}

	err = service.Delete(paste.ID)
	if err != nil {
		t.Errorf("Failed to delete a paste: %v", err)
	}

	paste, _ = service.Get(paste.ID)
	if paste != nil {
		t.Errorf("Found a paste after delete: %v", paste)
	}
}

func Test_DeleteNotFound(t *testing.T) {
	t.Parallel()
	err := service.Delete(0)
	if err != nil {
		t.Errorf("Should delete non-existing paste without error, got %v", err)
	}
}

func Test_List(t *testing.T) {
	t.Parallel()
	var p = createTestPaste()
	p.UserID = 1
	if _, err := service.Create(*p); err != nil {
		t.Errorf("failed to create a paste: %v", err)
		return
	}

	list, _ := service.List(1)
	if len(list) != 1 {
		t.Errorf("Expected a list of 1, got %d", len(list))
		return
	}
	if p.Title != list[0].Title {
		t.Errorf("wrong title, want %s got %s", p.Title, list[0].Title)
	}
}
