package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/src/store"
)

var svc *Service

func TestMain(m *testing.M) {
	svc = NewWithMemDB()
	os.Exit(m.Run())
}

// Test new paste
func TestNewPaste(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if p.ID == 0 {
		t.Error("expect paste to have an id")
	}
}

func TestNewPasteEmptyBody(t *testing.T) {
	t.Parallel()

	_, err := svc.NewPaste(PasteRequest{})
	if err == nil {
		t.Fatal("expected paste creation to fail")
	}
	if !errors.Is(err, ErrEmptyBody) {
		t.Errorf("expected error to be [%v], got [%v]", ErrEmptyBody, err)
	}
}

func TestNewPasteEmptyPrivacy(t *testing.T) {
	t.Parallel()

	_, err := svc.NewPaste(PasteRequest{Body: "Test body"})
	if err == nil {
		t.Fatal("expected paste creation to fail")
	}
	if !errors.Is(err, ErrWrongPrivacy) {
		t.Errorf("expected error to be [%v], got [%v]", ErrWrongPrivacy, err)
	}
}

func TestNewPasteWithPassword(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Body:     "Test body",
		Privacy:  "public",
		Password: "password",
	})
	if err != nil {
		t.Fatal("failed to create new paste")
	}
	if p.Password == "" || p.Password == "password" {
		t.Errorf("expected password to be hashed, got %s", p.Password)
	}
}

func TestNewPasteWithUser(t *testing.T) {
	t.Parallel()

	u := store.User{
		ID:   "test_user",
		Name: "Test User",
	}
	u, err := svc.GetOrUpdateUser(u)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		UserID:  u.ID,
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if p.User.ID != u.ID {
		t.Errorf("expect paste to have user id [%v], got [%v]", u.ID, p.User.ID)
	}
}

func TestNewPasteWithFakeUser(t *testing.T) {
	t.Parallel()

	_, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		UserID:  "non_existing_user",
	})
	if err == nil {
		t.Fatal("expected paste creation to fail")
	}
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected error to be [%v], got [%v]", ErrUserNotFound, err)
	}
}

func TestNewPasteWithExpirationMinutes(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "10m",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if time.Until(p.Expires) > 10*time.Minute {
		t.Errorf("expected paste expiration to be less than 10 minutes, got %v", p.Expires)
	}
}

func TestNewPasteWithExpirationHours(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "3h",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if time.Until(p.Expires) > 3*time.Hour {
		t.Errorf("expected paste expiration to be less than 3 hours, got %v", p.Expires)
	}
}

func TestNewPasteWithExpirationDays(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "5d",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if time.Until(p.Expires) > 5*24*time.Hour {
		t.Errorf("expected paste expiration to be less than 5 days, got %v", p.Expires)
	}
}

func TestNewPasteWithExpirationWeeks(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "2w",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if time.Until(p.Expires) > 14*24*time.Hour {
		t.Errorf("expected paste expiration to be less than 2 weeks, got %v", p.Expires)
	}
}

func TestNewPasteWithExpirationMonths(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "6M",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if time.Until(p.Expires.AddDate(0, -6, 0)) > time.Second {
		t.Errorf("expected paste expiration to be less than 6 months, got %v", p.Expires)
	}
}

func TestNewPasteWithExpirationYears(t *testing.T) {
	t.Parallel()

	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "2y",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	if time.Until(p.Expires.AddDate(-2, 0, 0)) > time.Second {
		t.Errorf("expected paste expiration to be less than 2 years, got %v", p.Expires)
	}
}

func TestNewPasteWrongExpiration(t *testing.T) {
	t.Parallel()

	_, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "abcdefg",
	})
	if err == nil {
		t.Fatal("expecte paste creation to fail")
	}
	if !errors.Is(err, ErrWrongDuration) {
		t.Errorf("expected error to be [%v], got [%v]", ErrWrongDuration, err)
	}

	_, err = svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
		Expires: "12g",
	})
	if err == nil {
		t.Fatal("expecte paste creation to fail")
	}
	if !errors.Is(err, ErrWrongDuration) {
		t.Errorf("expected error to be [%v], got [%v]", ErrWrongDuration, err)
	}
}

// Test get paste
func TestGetPaste(t *testing.T) {
	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "public",
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}

	paste, err := svc.GetPaste(p.URL(), "", "")
	if err != nil {
		t.Fatalf("failed to get the paste: %v", err)
	}
	if p.Title != paste.Title {
		t.Errorf("expected paste titles to be equal, want [%s], got [%s]", p.Title, paste.Title)
	}
}

func TestGetPasteWrongURL(t *testing.T) {
	_, err := svc.GetPaste("QwE-AsD", "", "")
	if err == nil {
		t.Fatalf("expected GetPaste to fail")
	}
}

func TestGetPasteDontExist(t *testing.T) {
	t.Parallel()

	_, err := svc.GetPaste("QwEAsD12", "", "")
	if err == nil {
		t.Fatal("expected GetPaste to fail")
	}
	if !errors.Is(err, ErrPasteNotFound) {
		t.Errorf("expected error to be [%v], got [%v]", ErrPasteNotFound, err)
	}
}

func TestGetPastePrivate(t *testing.T) {
	t.Parallel()

	u := store.User{
		ID:   "test_user_2",
		Name: "Test User",
	}
	u, err := svc.GetOrUpdateUser(u)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	p, err := svc.NewPaste(PasteRequest{
		Title:   "Test title",
		Body:    "Test body",
		Privacy: "private",
		UserID:  u.ID,
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	_, err = svc.GetPaste(p.URL(), u.ID, "")
	if err != nil {
		t.Errorf("expected to get private paste, got [%v]", err)
	}

	_, err = svc.GetPaste(p.URL(), "", "")
	if err == nil {
		t.Error("expected GetPaste to fail")
	}
	if !errors.Is(err, ErrPasteIsPrivate) {
		t.Errorf("expected error to be [%v], got [%v]", ErrPasteIsPrivate, err)
	}
}

func TestGetPasteWithPassword(t *testing.T) {
	t.Parallel()

	u := store.User{
		ID:   "test_user_3",
		Name: "Test User",
	}
	u, err := svc.GetOrUpdateUser(u)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	p, err := svc.NewPaste(PasteRequest{
		Title:    "Test title",
		Body:     "Test body",
		Privacy:  "public",
		Password: "password",
		UserID:   u.ID,
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	_, err = svc.GetPaste(p.URL(), "", "password")
	if err != nil {
		t.Errorf("expected to get paste with password, got [%v]", err)
	}

	_, err = svc.GetPaste(p.URL(), "", "")
	if err == nil {
		t.Error("expected GetPaste with password to fail")
	}
	if !errors.Is(err, ErrPasteHasPassword) {
		t.Errorf("expected error to be [%v], got [%v]", ErrPasteHasPassword, err)
	}

	_, err = svc.GetPaste(p.URL(), "", "12345")
	if err == nil {
		t.Error("expected GetPaste with password to fail")
	}
	if !errors.Is(err, ErrWrongPassword) {
		t.Errorf("expected error to be [%v], got [%v]", ErrWrongPassword, err)
	}
}

func TestGetPasteDeleteAfterRead(t *testing.T) {
	p, err := svc.NewPaste(PasteRequest{
		Title:           "Test title",
		Body:            "Test body",
		Privacy:         "public",
		DeleteAfterRead: true,
	})
	if err != nil {
		t.Fatalf("failed to create new paste: %v", err)
	}
	// Get the paste for the first time
	paste, err := svc.GetPaste(p.URL(), "", "")
	if err != nil {
		t.Fatalf("failed to get the paste: %v", err)
	}
	if p.Title != paste.Title {
		t.Errorf("expected paste titles to be equal, want [%s], got [%s]", p.Title, paste.Title)
	}
	// Try to get the paste again and check that it doesn't exist anymore
	paste, err = svc.GetPaste(p.URL(), "", "")
	if err == nil {
		t.Fatalf("expected paste to be deleted, got [%+v]", paste)
	}
	if !errors.Is(err, ErrPasteNotFound) {
		t.Errorf("expected error to be [%v], got [%v]", ErrPasteNotFound, err)
	}
}

// Test user pastes
func TestGetUserPastes(t *testing.T) {
	t.Parallel()

	u := store.User{
		ID:   "test_user_4",
		Name: "Test User",
	}
	u, err := svc.GetOrUpdateUser(u)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	for i := 0; i < 12; i++ {
		_, err := svc.NewPaste(PasteRequest{
			Body:    "Test body",
			Privacy: "public",
			UserID:  u.ID,
		})
		if err != nil {
			t.Fatalf("failed to create paste: %v", err)
		}
	}

	pastes, err := svc.UserPastes(u.ID)
	if err != nil {
		t.Errorf("failed to get user pastes: %v", err)
	}
	if len(pastes) != 10 {
		t.Errorf("expected to get 10 pastes, got %d", len(pastes))
	}
}
