package store

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/peterbourgon/diskv/v3"
)

var ErrNoUserID = fmt.Errorf("user must have a unique ID or Name")

const defaultDirMode = 0o755

// DiskConfig is the input configuration for disk storage of pastes.
type DiskConfig struct {
	// DataDir must be a writsable director for storing pastes and users.
	DataDir string `long:"data-dir" env:"DATA_DIR" default:"./data" description:"directory where pastes are stored"`
	// How much memory to use for k/v caches. This is x3 (3 caches). 0 is probably good for this app.
	CacheSize uint64 `long:"cache-size" env:"CACHE_SIZE" description:"file system storage cache size"`
	// The file mode given to new folders. Uses a sane default it omitted.
	DirMode os.FileMode `long:"dir-mode" env:"DIR_MODE" description:"file mode for new directories"`
}

// DiskStore satisfies the main paste Interface.
type DiskStore struct {
	users      *diskv.Diskv
	pastes     *diskv.Diskv
	userPastes *diskv.Diskv
	pasteCount int64
	userList   map[string]struct{} // we only use this for counts, but it could be expanded.
	expiring   chan Paste
	sync.RWMutex
}

// Fail if the struct does not match the Interface.
var _ = Interface(&DiskStore{})

// NewDiskStorage should be called once on startup to initialize a disk storage backend for pastes.
func NewDiskStorage(config *DiskConfig) (*DiskStore, error) {
	if err := makeDiskStorageFolders(config); err != nil {
		return nil, err
	}

	store := &DiskStore{
		userList: make(map[string]struct{}),
		expiring: make(chan Paste),
		users: diskv.New(diskv.Options{
			BasePath:     filepath.Join(config.DataDir, "users"),
			CacheSizeMax: config.CacheSize,
		}),
		pastes: diskv.New(diskv.Options{
			BasePath:     filepath.Join(config.DataDir, "pastes"),
			CacheSizeMax: config.CacheSize,
		}),
		userPastes: diskv.New(diskv.Options{
			BasePath:     filepath.Join(config.DataDir, "user_pastes"),
			CacheSizeMax: config.CacheSize,
		}),
	}

	go store.cleanExpired()
	defer store.fillCaches()

	return store, nil
}

func (f *DiskStore) cleanExpired() {
	expiring := make(map[string]time.Time)

	// Store all expiring paste IDs in memory on startup.
	// This may be slow on systems with a lot of pastes.
	// Should be okay though since this runs in a go routine.
	for key := range f.pastes.Keys(nil) {
		var paste Paste
		if _ = f.getFromDisk(f.pastes, key, &paste); !paste.Expires.IsZero() {
			expiring[key] = paste.Expires
		}
	}

	// Check every minute if something is expiring.
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			for pasteID, when := range expiring {
				if when.Before(now) {
					paste := Paste{}
					_ = f.getFromDisk(f.pastes, pasteID, &paste)
					_ = f.delete(paste) // should log this error.
					delete(expiring, pasteID)
				}
			}
		case paste, ok := <-f.expiring:
			if !ok {
				return // f.expiring can be closed to exit this loop.
			} else if !paste.Expires.IsZero() {
				expiring[f.intStr(paste.ID)] = paste.Expires
			}
		}
	}
}

func makeDiskStorageFolders(config *DiskConfig) error {
	if config.DirMode == 0 {
		config.DirMode = defaultDirMode
	}

	dirStat, err := os.Stat(config.DataDir)
	if err != nil {
		return fmt.Errorf("data dir missing? %w", err)
	}

	if !dirStat.IsDir() {
		return fmt.Errorf("data dir is not a directory: %s", dirStat.Name())
	}

	// Stores user info only.
	err = os.MkdirAll(filepath.Join(config.DataDir, "users"), config.DirMode)
	if err != nil {
		return fmt.Errorf("creating users data store: %w", err)
	}

	// Stores all paste data, along with the user that made it.
	err = os.MkdirAll(filepath.Join(config.DataDir, "pastes"), config.DirMode)
	if err != nil {
		return fmt.Errorf("creating pastes data store: %w", err)
	}

	// Stores a list of users with list of paste IDs.
	// This is so we do not have to iterate every paste to get a user's list.
	err = os.MkdirAll(filepath.Join(config.DataDir, "user_pastes"), config.DirMode)
	if err != nil {
		return fmt.Errorf("creating user-pastes data store: %w", err)
	}

	return nil
}

// Totals return total counts for pastes and users.
func (f *DiskStore) Totals() (int64, int64) {
	f.RLock()
	defer f.RUnlock()

	return f.pasteCount, int64(len(f.userList))
}

// Create new paste and return its id.
func (f *DiskStore) Create(paste Paste) (int64, error) {
	paste.ID = paste.CreatedAt.UnixNano()

	if err := f.writePaste(paste); err != nil {
		return 0, err
	}

	f.expiring <- paste

	f.Lock()
	defer f.Unlock()
	f.pasteCount++

	return paste.ID, nil
}

func (f *DiskStore) writePaste(paste Paste) error {
	if err := f.saveToDisk(f.pastes, f.intStr(paste.ID), &paste); err != nil {
		return err
	}

	if err := f.writeUsersPaste(paste); err != nil {
		return fmt.Errorf("writing user-paste: %w", err)
	}

	return nil
}

func (f *DiskStore) writeUsersPaste(paste Paste) error {
	if paste.User.ID == "" {
		return nil
	}

	pasteList := make(map[int64]struct{})
	_ = f.getFromDisk(f.userPastes, paste.User.ID, &pasteList)
	pasteList[paste.ID] = struct{}{}

	return f.saveToDisk(f.userPastes, paste.User.ID, &pasteList)
}

// Delete paste by id.
func (f *DiskStore) Delete(pasteID int64) error {
	paste, _ := f.Get(pasteID)
	return f.delete(paste)
}

func (f *DiskStore) delete(paste Paste) error {
	if paste.ID == 0 {
		return nil
	}

	err := f.pastes.Erase(f.intStr(paste.ID))
	if err != nil {
		return fmt.Errorf("disk.Delete: %w", err)
	}

	f.Lock()
	f.pasteCount--
	f.Unlock()

	if paste.User.ID != "" {
		return f.deleteUserPaste(paste)
	}

	return nil
}

func (f *DiskStore) deleteUserPaste(paste Paste) error {
	ikeys := make(map[int64]struct{})

	err := f.getFromDisk(f.userPastes, paste.User.ID, &ikeys)
	if err != nil {
		return fmt.Errorf("disk.Delete (user-paste): %w", err)
	}

	delete(ikeys, paste.ID)

	err = f.saveToDisk(f.userPastes, paste.User.ID, &ikeys)
	if err != nil {
		return fmt.Errorf("disk.Delete (save user-paste): %w", err)
	}

	return nil
}

// Find pastes.
func (f *DiskStore) Find(req FindRequest) ([]Paste, error) {
	var (
		keys   = []string{}
		pastes = []Paste{}
	)

	if req.UserID != "" {
		ikeys := make(map[int64]struct{})

		err := f.getFromDisk(f.userPastes, req.UserID, &ikeys)
		if err != nil {
			return pastes, nil //nolint:nilerr // user has no pastes, do not return an error
		}

		for ikey := range ikeys {
			keys = append(keys, f.intStr(ikey))
		}
	} else {
		for key := range f.pastes.Keys(nil) {
			keys = append(keys, key)
		}
	}

	var sortCreated bool
	if req.Sort == "+created" {
		sortCreated = true
		sort.Strings(keys)
	} else if req.Sort == "-created" {
		sortCreated = true
		sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	}

	for _, key := range keys {
		if sortCreated && req.Limit != 0 && len(pastes) >= req.Limit {
			break
		}

		var paste Paste
		if err := f.getFromDisk(f.pastes, key, &paste); err != nil {
			return nil, fmt.Errorf("disk.Find: %w", err)
		}

		if filterPaste(req, paste) {
			pastes = append(pastes, paste)
		}
	}

	if !sortCreated {
		sortPastes(req, pastes)
	}

	if sortCreated && req.Limit > 0 {
		return pastes, nil
	}

	// Slice with skip and limit
	return limitPastes(req, pastes), nil
}

// Count return pastes count for a user.
func (f *DiskStore) Count(req FindRequest) int64 {
	if req.UserID == "" {
		f.RLock()
		defer f.RUnlock()

		return f.pasteCount
	}

	pasteList := make(map[int64]struct{})
	if err := f.getFromDisk(f.userPastes, req.UserID, &pasteList); err != nil {
		return 0
	}

	return int64(len(pasteList))
}

// Get paste by id.
func (f *DiskStore) Get(pasteID int64) (Paste, error) {
	var paste Paste
	if err := f.getFromDisk(f.pastes, f.intStr(pasteID), &paste); err != nil {
		return paste, fmt.Errorf("disk.Get: %w", err)
	}

	if !paste.Expires.IsZero() && paste.Expires.Before(time.Now()) {
		_ = f.Delete(pasteID)
		return Paste{}, fmt.Errorf("paste expired")
	}

	return paste, nil
}

// Update paste information and return updated paste.
func (f *DiskStore) Update(paste Paste) (Paste, error) {
	if existing, err := f.Get(paste.ID); err != nil {
		return existing, err
	}

	if err := f.writePaste(paste); err != nil {
		return paste, err
	}

	f.expiring <- paste

	return paste, nil
}

// SaveUser creates or updates a user.
func (f *DiskStore) SaveUser(user User) (string, error) {
	if err := f.saveToDisk(f.users, user.ID, &user); err != nil {
		return "", fmt.Errorf("disk.SaveUser: %w", err)
	}

	f.Lock()
	defer f.Unlock()
	f.userList[user.ID] = struct{}{}

	return user.ID, nil
}

// User returns a user by id.
func (f *DiskStore) User(userID string) (User, error) {
	var user User
	if err := f.getFromDisk(f.users, userID, &user); err != nil {
		return user, fmt.Errorf("disk.User: %w", err)
	}

	return user, nil
}

// fillCaches stores the user list and paste count in memory.
// This should only run once on startup.
// The data is appended-to and updated as the app runs.
func (f *DiskStore) fillCaches() {
	f.Lock()
	defer f.Unlock()

	for username := range f.users.Keys(nil) {
		f.userList[username] = struct{}{}
	}

	for range f.pastes.Keys(nil) {
		f.pasteCount++
	}
}

func (f *DiskStore) saveToDisk(disk *diskv.Diskv, storeID string, data interface{}) error {
	var (
		buf bytes.Buffer
		enc = gob.NewEncoder(&buf)
	)

	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encoding buffer: %w", err)
	}

	if err := disk.WriteStream(storeID, &buf, true); err != nil {
		return fmt.Errorf("writing data: %w", err)
	}

	return nil
}

func (f *DiskStore) getFromDisk(disk *diskv.Diskv, storeID string, data interface{}) error {
	buf, err := disk.ReadStream(storeID, true)
	if err != nil {
		return fmt.Errorf("reading storage (id:%s): %w", storeID, err)
	}
	defer buf.Close()

	if err := gob.NewDecoder(buf).Decode(data); err != nil {
		return fmt.Errorf("decoding storage buffer: %w", err)
	}

	return nil
}

// Our library uses an int64 for paste IDs,
// but the disk storage library uses strings for keys.
// This procedure handles the conversion.
func (f *DiskStore) intStr(pasteID int64) string {
	const base10 = 10
	return strconv.FormatInt(pasteID, base10)
}
