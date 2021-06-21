package memory

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/api"
)

var service = New()

// createTestPaste create a paste with random ID and Body
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

func Test_Create(t *testing.T) {
	var p = createTestPaste()
	if err := service.Create(p); err != nil {
		t.Errorf("Failed to create a paste: %v", err)
	}
}

func Test_CreateAlreadyExists(t *testing.T) {
	var p = createTestPaste()
	if err := service.Create(p); err != nil {
		t.Errorf("Failed to create a paste: %v", err)
	}
	if err := service.Create(p); err == nil {
		t.Errorf("Created a paste with existing ID: %v", p)
	}
}

func Test_Paste(t *testing.T) {
	var p = createTestPaste()
	service.Create(p)

	_, err := service.Paste(p.ID)
	if err != nil {
		t.Errorf("Failed to find a paste: %v", err)
	}
}

func Test_PasteNotFound(t *testing.T) {
	_, err := service.Paste("NotFound")
	if err == nil {
		t.Error("No error for non-existing paste")
	} else {
		t.Logf("%v", err)
	}
}

func Test_Delete(t *testing.T) {
	var p = createTestPaste()
	service.Create(p)

	err := service.Delete(p.ID)
	if err != nil {
		t.Errorf("Failed to delete a paste: %v", err)
	}

	paste, err := service.Paste(p.ID)
	if err == nil {
		t.Errorf("Found a paste after delete: %v", paste)
	}
}

func Test_DeleteNotFound(t *testing.T) {
	err := service.Delete("NotFound")
	if err == nil {
		t.Error("No error for non-existing paste")
	} else {
		t.Logf("%v", err)
	}
}
