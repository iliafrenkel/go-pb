package api

import "time"

type Paste struct {
	ID      string
	Title   string
	Body    []byte
	Expires time.Time
}

type PasteService interface {
	Paste(id string) (*Paste, error)
	Create(p Paste) error
	Delete(id string) error
}
