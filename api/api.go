package api

import (
	"time"

	"github.com/iliafrenkel/go-pb/api/base62"
)

// Paste is a the type that represents a single paste.
type Paste struct {
	ID      uint64    `json:"id"`
	Title   string    `json:"title"`
	Body    []byte    `json:"body"`
	Expires time.Time `json:"expires"`
}

func (p *Paste) URL() string {
	return base62.Encode(p.ID)
}

// PasteService is the interface that defines methods for working with Pastes.
//
// Implementations should define the underlying storage such as database,
// plain files or even memory.
type PasteService interface {
	Paste(id uint64) (*Paste, error)
	Create(p *Paste) error
	Delete(id uint64) error
}
