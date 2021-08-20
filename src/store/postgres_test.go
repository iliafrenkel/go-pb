package store

import (
	"sort"
	"testing"
	"time"
)

// TestDelete tests that we can delete a paste.
func TestDeletePDB(t *testing.T) {
	t.Parallel()
	// Create random paste
	paste := randomPaste(randomUser())
	id, err := pdb.Create(paste)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}
	// Delete the paste and check that it was indeed deleted.
	err = pdb.Delete(id)
	if err != nil {
		t.Fatalf("failed to delete paste: %v", err)
	}
	p, err := pdb.Get(id)
	if err != nil {
		t.Fatalf("failed to get paste: %v", err)
	}
	if p != (Paste{}) {
		t.Errorf("expected paste to be deleted but found %+v", p)
	}
}

// TestDelete tests that we can't delete a paste if it doesn't exist.
func TestDeleteNonExistingPDB(t *testing.T) {
	t.Parallel()
	// Create random paste
	paste := randomPaste(randomUser())
	// Delete the paste and check that it was indeed deleted.
	err := pdb.Delete(paste.ID)
	if err == nil {
		t.Fatalf("expected delete to fail")
	}
}

// TestFind tests that we can find a paste using various parameters.
func TestFindPDB(t *testing.T) {
	t.Parallel()
	// Create 2 users with 10 pastes each and 10 anonymous pastes
	usr1 := randomUser()
	usr1.ID = "find_user_1"
	usr2 := randomUser()
	usr2.ID = "find_user_2"
	for i := 0; i < 10; i++ {
		p1 := randomPaste(usr1)
		p1.CreatedAt = time.Now().AddDate(0, 0, -1*i)
		p1.Expires = time.Now().AddDate(0, 1*i, 0)
		p1.Views = int64(10*i + 1)
		pdb.Create(p1)
		p2 := randomPaste(usr2)
		p2.CreatedAt = time.Now().AddDate(0, 0, -1*i)
		pdb.Create(p2)
		p3 := randomPaste(User{})
		p3.CreatedAt = time.Now().AddDate(0, 0, -1*i)
		pdb.Create(p3)
	}

	for _, tc := range findTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pastes, err := pdb.Find(FindRequest{
				UserID: tc.uid,
				Sort:   tc.sort,
				Limit:  tc.limit,
				Skip:   tc.skip,
			})
			if err != nil {
				t.Fatalf("failed to find pastes: %v", err)
			}
			if len(pastes) != tc.exp {
				t.Errorf("expected to find %d pastes, got %d", tc.exp, len(pastes))
			}
			if tc.sort == "" {
				return
			}
			if !sort.SliceIsSorted(pastes, func(i, j int) bool {
				switch tc.sort {
				case "-created":
					return pastes[i].CreatedAt.After(pastes[j].CreatedAt)
				case "+created":
					return pastes[i].CreatedAt.Before(pastes[j].CreatedAt)
				case "-expires":
					return pastes[i].Expires.After(pastes[j].Expires)
				case "+expires":
					return pastes[i].Expires.Before(pastes[j].Expires)
				case "-views":
					return pastes[i].Views > pastes[j].Views
				case "+views":
					return pastes[i].Views < pastes[j].Views
				default:
					return false
				}
			}) {
				t.Errorf("expected pastes to be sorted by %s, got %+v", tc.sort, pastes)
			}
		})
	}
}

func TestUpdatePDB(t *testing.T) {
	t.Parallel()
	usr := randomUser()
	// Create random paste
	paste := randomPaste(usr)
	id, err := pdb.Create(paste)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}
	// Update the paste
	paste, _ = pdb.Get(id)
	paste.Views = 42
	p, _ := pdb.Update(paste)

	if p.ID != id {
		t.Errorf("expected paste to have the same id [%d], got [%d]", id, p.ID)
	}

	if p.Views != paste.Views {
		t.Errorf("expected paste views to be updated to [%d], got [%d]", paste.Views, p.Views)
	}
}

func TestUpdateNonExistingPDB(t *testing.T) {
	t.Parallel()
	usr := randomUser()
	// Create random paste
	paste := randomPaste(usr)
	p, err := pdb.Update(paste)

	if err == nil {
		t.Error("expected paste update to fail")
	}
	if p != (Paste{}) {
		t.Errorf("expected paste to be empty, got [%+v]", p)
	}
}

func TestGetUserPDB(t *testing.T) {
	t.Parallel()
	usr := randomUser()
	id, err := pdb.SaveUser(usr)
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}
	u, err := pdb.User(id)
	if err != nil {
		t.Errorf("user not found: %v", err)
	}
	if usr.ID != u.ID || usr.Name != u.Name {
		t.Errorf("expected user to be saved as [%+v], got [%+v]", usr, u)
	}
}

func TestGetUserNotExistingPDB(t *testing.T) {
	t.Parallel()
	usr := randomUser()
	u, err := pdb.User(usr.ID)
	if err == nil {
		t.Errorf("expected user to be not found")
	}
	if u != (User{}) {
		t.Errorf("expected user to be empty, got %+v", u)
	}

}
