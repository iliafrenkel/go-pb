// The memory package provides methods to work with pastes using memory as a
// storage.
package memory

import (
	"errors"

	"github.com/iliafrenkel/go-pb/api"
)

type PasteService struct {
	Pastes map[string]*api.Paste
}

func New() *PasteService {
	var s PasteService
	s.Pastes = make(map[string]*api.Paste)
	return &s
}

// Paste returns a paste by it's ID.
func (s *PasteService) Paste(id string) (*api.Paste, error) {
	if p, ok := s.Pastes[id]; ok {
		return p, nil
	}
	return nil, errors.New("Paste not found")
}

// Create adds a new paste to the storage
func (s *PasteService) Create(p *api.Paste) error {
	if _, ok := s.Pastes[p.ID]; ok {
		return errors.New("Paste ID already exists")
	}
	s.Pastes[p.ID] = p

	return nil
}

// Delete removes the paste from the storage
func (s *PasteService) Delete(id string) error {
	if _, ok := s.Pastes[id]; !ok {
		return errors.New("Paste not found")
	}

	delete(s.Pastes, id)

	return nil
}
