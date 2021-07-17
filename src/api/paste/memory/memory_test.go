package memory

import (
	"testing"

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

func Test_GetPaste(t *testing.T) {
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
	_, err := service.Get(0)
	if err == nil {
		t.Error("No error for non-existing paste")
	} else {
		t.Logf("%v", err)
	}
}

func Test_Delete(t *testing.T) {
	var p = createTestPaste()
	paste, err := service.Create(*p)
	if err != nil {
		t.Errorf("failed to create a paste: %v", err)
	}

	err = service.Delete(paste.ID)
	if err != nil {
		t.Errorf("Failed to delete a paste: %v", err)
	}

	paste, err = service.Get(paste.ID)
	if err == nil {
		t.Errorf("Found a paste after delete: %v", paste)
	}
}

func Test_DeleteNotFound(t *testing.T) {
	err := service.Delete(0)
	if err == nil {
		t.Error("No error for non-existing paste")
	} else {
		t.Logf("%v", err)
	}
}
