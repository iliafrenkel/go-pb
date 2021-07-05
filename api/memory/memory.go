/* The memory package provides methods to work with pastes using memory as a
 * storage.
 *
 * This package provides a PasteService type that implements api.PasteService
 * interface and use a map of Pastes as a storage.
 *
 * Note: according to the documentation (https://blog.golang.org/maps#TOC_6.),
 * maps are not safe for concurrent use.
 */
package memory

import (
	"errors"

	"github.com/iliafrenkel/go-pb/api"
)

// PasteService stores all the pastes in memory and implements the
// api.PasteService interface.
type PasteService struct {
	Pastes map[uint64]*api.Paste
}

// New returns new PasteService with an empty map of pastes.
func New() *PasteService {
	var s PasteService
	s.Pastes = make(map[uint64]*api.Paste)
	return &s
}

// Paste returns a paste by it's ID.
func (s *PasteService) Paste(id uint64) (*api.Paste, error) {
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
func (s *PasteService) Delete(id uint64) error {
	if _, ok := s.Pastes[id]; !ok {
		return errors.New("Paste not found")
	}

	delete(s.Pastes, id)

	return nil
}
