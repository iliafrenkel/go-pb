package store

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	mdb     *MemDB
	pdb     *PostgresDB
	ddb     *DiskStore
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

// testCaseForFind used by TestFind test
type testCaseForFind struct {
	name    string // test case name
	uid     string // user id
	sort    string // sort direction
	limit   int    // max records
	skip    int    // records to skip
	privacy string // filter by paste privacy
	exp     int    // expected result length
}

var findTestCases = []testCaseForFind{
	{
		name:  "All user pastes",
		uid:   "find_user_1",
		sort:  "",
		limit: 11,
		skip:  0,
		exp:   10,
	}, {
		name:  "Check limit",
		uid:   "find_user_2",
		sort:  "",
		limit: 5,
		skip:  0,
		exp:   5,
	}, {
		name:  "Check skip",
		uid:   "find_user_2",
		sort:  "",
		limit: 5,
		skip:  6,
		exp:   4,
	}, {
		name:  "Check skip over limit",
		uid:   "find_user_2",
		sort:  "",
		limit: 5,
		skip:  12,
		exp:   0,
	}, {
		name:  "Check sort by -created",
		uid:   "find_user_1",
		sort:  "-created",
		limit: 10,
		skip:  0,
		exp:   10,
	}, {
		name:  "Check sort by +created",
		uid:   "find_user_1",
		sort:  "+created",
		limit: 10,
		skip:  0,
		exp:   10,
	}, {
		name:  "Check sort by -expires",
		uid:   "find_user_1",
		sort:  "-expires",
		limit: 10,
		skip:  0,
		exp:   10,
	}, {
		name:  "Check sort by +expires",
		uid:   "find_user_1",
		sort:  "+expires",
		limit: 10,
		skip:  0,
		exp:   10,
	}, {
		name:  "Check sort by -views",
		uid:   "find_user_1",
		sort:  "-views",
		limit: 10,
		skip:  0,
		exp:   10,
	}, {
		name:  "Check sort by +views",
		uid:   "find_user_1",
		sort:  "+views",
		limit: 10,
		skip:  0,
		exp:   10,
	}, {
		name:    "Check public",
		uid:     "",
		sort:    "-created",
		limit:   20,
		skip:    0,
		privacy: "public",
		exp:     20,
	},
}

// TestMain is a setup function for the test suite. It creates a new MemDB
// instance and seeds random generator.
func TestMain(m *testing.M) {
	mdb = NewMemDB()

	dir, err := os.MkdirTemp("", "go-pb-tests")
	if err != nil {
		fmt.Printf("got error making disk store folder: %s\n", err)
		os.Exit(1)
	}

	ddb, err = NewDiskStorage(&DiskConfig{DataDir: dir})
	if err != nil {
		fmt.Printf("got error making disk store: %s\n", err)
		os.Exit(1)
	}

	pdb, err = NewPostgresDB("host=localhost user=test password=test dbname=test port=5432 sslmode=disable", true)
	pdb.db.AllowGlobalUpdate = true
	pdb.db.Delete(Paste{})
	pdb.db.Delete(User{})
	if err != nil {
		fmt.Printf("Failed to create a PostgresDB store: %v\n", err)
		os.Exit(1)
	}
	rand.Seed(time.Now().UnixNano())

	c := m.Run()

	pdb.db.Delete(Paste{})
	pdb.db.Delete(User{})

	os.RemoveAll(dir)
	os.Exit(c)
}

// randSeq generates random string of a given size.
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// randomUser creates a User with random values.
func randomUser() User {
	u := User{
		ID:    randSeq(10),
		Name:  randSeq(10),
		Email: "",
		IP:    "",
		Admin: false,
	}
	return u
}

// randomPaste creates a Paste with random values.
func randomPaste(usr User) Paste {
	p := Paste{
		ID:              rand.Int63(),
		Title:           randSeq(10),
		Body:            randSeq(10),
		Expires:         time.Time{},
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "",
		CreatedAt:       time.Now(),
		Syntax:          "none",
		UserID:          usr.ID,
		User:            usr,
		Views:           0,
	}
	return p
}
func TestPasteURL(t *testing.T) {
	t.Parallel()

	p := Paste{
		ID:              123,
		Title:           "",
		Body:            "qwe",
		Expires:         time.Time{},
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "",
		CreatedAt:       time.Now(),
		Syntax:          "text",
		User:            User{},
		Views:           0,
	}

	id, _ := p.URL2ID(p.URL())
	if p.ID != id {
		t.Errorf("expected paste id to be %d, got %d", p.ID, id)
	}

	_, err := p.URL2ID("@#%$#")
	if err == nil {
		t.Error("expected decoding to fail")
	}
}

func TestPasteExpiration(t *testing.T) {
	t.Parallel()

	p := Paste{
		ID:              123,
		Title:           "",
		Body:            "qwe",
		Expires:         time.Now().Add(30 * time.Second),
		DeleteAfterRead: false,
		Privacy:         "public",
		Password:        "",
		CreatedAt:       time.Now(),
		Syntax:          "text",
		User:            User{},
		Views:           0,
	}
	if !strings.HasSuffix(p.Expiration(), "sec") {
		t.Errorf("expected expiration to have [sec], got [%s]", p.Expiration())
	}

	p.Expires = time.Now().Add(11 * time.Minute)
	if !strings.HasSuffix(p.Expiration(), "min") {
		t.Errorf("expected expiration to have [min], got [%s]", p.Expiration())
	}

	p.Expires = time.Now().Add(13 * time.Hour)
	if p.Expiration()[2:3] != ":" && p.Expiration()[5:6] != ":" {
		t.Errorf("expected expiration to be [13:00:00], got [%s]", p.Expiration())
	}

	p.Expires = time.Now().Add(96 * time.Hour)
	if !strings.HasSuffix(p.Expiration(), "days") {
		t.Errorf("expected expiration to have [days], got [%s]", p.Expiration())
	}

	p.Expires = time.Now().AddDate(0, 5, 0)
	if !strings.HasSuffix(p.Expiration(), "months") {
		t.Errorf("expected expiration to have [months], got [%s]", p.Expiration())
	}

	p.Expires = time.Now().AddDate(2, 0, 0)
	if !strings.HasSuffix(p.Expiration(), "years") {
		t.Errorf("expected expiration to have [years], got [%s]", p.Expiration())
	}

	p.Expires = time.Time{}
	if p.Expiration() != "Never" {
		t.Errorf("expected expiration to be [Never], got [%s]", p.Expiration())
	}

	p.Expires = time.Now().Add(1*time.Second - 1*time.Millisecond)
	if !strings.HasPrefix(p.Expiration(), "999") || !strings.HasSuffix(p.Expiration(), "ms") {
		t.Errorf("expected expiration to be [999ms], got [%s]", p.Expiration())
	}
}
