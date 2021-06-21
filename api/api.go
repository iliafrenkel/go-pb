package api

import "time"

// Paste is a the type that represents a single paste.
type Paste struct {
	ID      string    `json:"id,string"`
	Title   string    `json:"title,string"`
	Body    []byte    `json:"body"`
	Expires time.Time `json:"expires"`
}

// PasteService is the interface that defines methods for working with Pastes.
//
// Implementations should define the underlying storage such as database,
// plain files or even memory.
type PasteService interface {
	Paste(id string) (*Paste, error)
	Create(p *Paste) error
	Delete(id string) error
}
