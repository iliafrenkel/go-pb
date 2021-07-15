package memory

import (
	"math/rand"
	"testing"
	"time"

	"github.com/iliafrenkel/go-pb/src/api"
	"github.com/iliafrenkel/go-pb/src/api/base62"
)

var service = New()

// createTestPaste create a paste with random ID and Body
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
	_, err := service.Paste(0)
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
	err := service.Delete(0)
	if err == nil {
		t.Error("No error for non-existing paste")
	} else {
		t.Logf("%v", err)
	}
}
