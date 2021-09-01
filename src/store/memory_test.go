package store

import (
	"math/rand"
	"sort"
	"testing"
	"time"
)

// TestTotals tests that we can count pastes and users correctly.
func TestTotals(t *testing.T) {
	t.Parallel()

	// We need a dedicated store because other test running in parallel
	// will affect the counts.
	m := NewMemDB()

	var usr User
	var paste Paste

	// Generate a bunch of users and pastes
	uCnt := rand.Int63n(10)
	pCnt := rand.Int63n(20)
	for i := int64(0); i < uCnt; i++ {
		usr = randomUser()
		_, err := m.SaveUser(usr)
		if err != nil {
			t.Fatalf("failed to save user: %v", err)
		}
		for j := int64(0); j < pCnt; j++ {
			u, err := m.User(usr.ID)
			if err != nil {
				t.Fatalf("failed to get user: %v", err)
			}
			paste = randomPaste(u)
			_, err = m.Create(paste)
			if err != nil {
				t.Fatalf("failed to create paste: %v", err)
			}
		}
	}

	// Check the counts
	wantUsers := uCnt
	wantPastes := uCnt * pCnt
	gotPastes, gotUsers := m.Totals()

	if wantUsers != gotUsers {
		t.Errorf("users count is incorrect, want %d, got %d", wantUsers, gotUsers)
	}
	if wantPastes != gotPastes {
		t.Errorf("pastes count is incorrect, want %d, got %d", wantPastes, gotPastes)
	}
}

func TestCount(t *testing.T) {
	t.Parallel()

	usr := randomUser()
	_, err := mdb.SaveUser(usr)
	if err != nil {
		t.Fatalf("failed to save user: %v", err)
	}
	pCnt := rand.Int63n(20)
	for i := int64(0); i < pCnt; i++ {
		paste := randomPaste(usr)
		_, err = mdb.Create(paste)
		if err != nil {
			t.Fatalf("failed to create paste: %v", err)
		}
	}
	got := mdb.Count(FindRequest{UserID: usr.ID})
	if got != pCnt {
		t.Errorf("pastes count for user %s is incorrect, want %d got %d", usr.ID, pCnt, got)
	}
	// count public
	got = mdb.Count(FindRequest{Privacy: "public"})
	if got < pCnt {
		t.Errorf("public pastes count is incorrect, expected %d to be greater than %d", got, pCnt)
	}
}

// TestDelete tests that we can delete a paste.
func TestDelete(t *testing.T) {
	t.Parallel()

	// Create random paste
	paste := randomPaste(User{})
	_, err := mdb.Create(paste)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}
	// Delete the paste and check that it was indeed deleted.
	err = mdb.Delete(paste.ID)
	if err != nil {
		t.Fatalf("failed to delete paste: %v", err)
	}
	p, err := mdb.Get(paste.ID)
	if err != nil {
		t.Fatalf("failed to get paste: %v", err)
	}
	if p != (Paste{}) {
		t.Errorf("expected paste to be deleted but found %+v", p)
	}
}

// TestFind tests that we can find a paste using various parameters.
func TestFind(t *testing.T) {
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
		p1.Views = int64(10 * i)
		mdb.Create(p1)
		p2 := randomPaste(usr2)
		p2.CreatedAt = time.Now().AddDate(0, 0, -1*i)
		mdb.Create(p2)
		p3 := randomPaste(User{})
		p3.CreatedAt = time.Now().AddDate(0, 0, -1*i)
		p3.Privacy = "private"
		mdb.Create(p3)
	}

	for _, tc := range findTestCases {
		t.Run(tc.name, func(t *testing.T) {
			pastes, err := mdb.Find(FindRequest{
				UserID:  tc.uid,
				Sort:    tc.sort,
				Limit:   tc.limit,
				Skip:    tc.skip,
				Privacy: tc.privacy,
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

func TestUpdate(t *testing.T) {
	t.Parallel()

	// Create random paste
	paste := randomPaste(User{})
	id, err := mdb.Create(paste)
	if err != nil {
		t.Fatalf("failed to create paste: %v", err)
	}
	// Update the paste
	paste, _ = mdb.Get(id)
	paste.Views = 42
	p, _ := mdb.Update(paste)

	if p.ID != id {
		t.Errorf("expected paste to have the same id [%d], got [%d]", id, p.ID)
	}

	if p.Views != paste.Views {
		t.Errorf("expected paste views to be updated to [%d], got [%d]", paste.Views, p.Views)
	}
}

func TestUpdateNonExisting(t *testing.T) {
	t.Parallel()

	// Create random paste
	paste := randomPaste(User{})
	p, _ := mdb.Update(paste)

	if p != (Paste{}) {
		t.Errorf("expected paste to be empty, got [%+v]", p)
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()
	usr := randomUser()
	id, err := mdb.SaveUser(usr)
	if err != nil {
		t.Errorf("failed to create user: %v", err)
	}
	u, err := mdb.User(id)
	if err != nil {
		t.Errorf("user not found: %v", err)
	}
	if usr.ID != u.ID || usr.Name != u.Name {
		t.Errorf("expected user to be saved as [%+v], got [%+v]", usr, u)
	}
}

func TestGetUserNotExisting(t *testing.T) {
	t.Parallel()
	usr := randomUser()
	u, err := mdb.User(usr.ID)
	if err == nil {
		t.Errorf("expected user to be not found")
	}
	if u != (User{}) {
		t.Errorf("expected user to be empty, got %+v", u)
	}

}
