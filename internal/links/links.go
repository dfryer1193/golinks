package links

import (
	"github.com/dfryer1193/golinks/internal/links/storage"
	"net/url"
	"sync"
)

type ParseError struct{}

// LinkMap houses the map of redirects, and keeps track of the backing file for
// maintaining the map across restarts. It also handles thread safety.
type LinkMap struct {
	store   storage.Storage
	m       map[string]string
	mapLock *sync.RWMutex
}

// NewLinkMap generates a new LinkMap object, with the requested config if it
// exists. If the requested config does not exist, it falls back to the default
// locations. If the default locations do not exist, the program will exit with
// an error.
func NewLinkMap(persistType storage.StorageType, requestedConfig string) *LinkMap {
	store := buildStorage(persistType, requestedConfig)

	linkMap := LinkMap{
		store:   store,
		m:       store.Read(),
		mapLock: &sync.RWMutex{},
	}

	go linkMap.handleReload()

	return &linkMap
}

func buildStorage(persistType storage.StorageType, requestedConfig string) storage.Storage {
	switch persistType {
	case storage.NONE:
		return storage.NewNoneStorage()
	case storage.FILE:
		return storage.NewFileStorage(requestedConfig)
	default:
		return storage.NewFileStorage("")
	}
}

func (l *LinkMap) handleReload() {
	reloadChannel := l.store.GetReloadChannel()
	// If we don't receive a reload channel, we will never receive updates, so we can just stop watching
	if reloadChannel == nil {
		return
	}

	for {
		select {
		case signal := <-reloadChannel:
			if signal {
				l.reload()
			}
		}
	}
}

func (l *LinkMap) reload() {
	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	l.m = l.store.Read()
}

// Get returns the url and state of existence for a single key.
func (l *LinkMap) Get(key string) (string, bool) {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()

	target, exists := l.m[key]
	return target, exists
}

// GetAll returns a map containing all of the entries from the current LinkMap object
func (l *LinkMap) GetAll() map[string]string {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()

	return l.m
}

func (l *LinkMap) GetAllKeys() []string {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()

	keys := make([]string, len(l.m))
	i := 0
	for key := range l.m {
		keys[i] = key
		i++
	}

	return keys
}

func (l *LinkMap) GetFiltered(keys []string) map[string]string {
	filteredMap := make(map[string]string, len(keys))

	l.mapLock.RLock()
	defer l.mapLock.RUnlock()
	for _, key := range keys {
		if v, exists := l.m[key]; exists {
			filteredMap[key] = v
		}
	}

	return filteredMap
}

// Put appends a new entry to the link map. If the entry already exists, it will
// be duplicated in the backing file, and the value in the live map will be
// replaced.
func (l *LinkMap) Put(key string, target *url.URL) error {
	go l.store.Put(key, target.String())

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	l.m[key] = target.String()

	return nil
}

// Delete removes an entry from the link map. If the key is not present in the
// map, this is a no-op
func (l *LinkMap) Delete(key string) error {
	// Can skip filesystem-intensive writes if the entry already doesn't exist
	l.mapLock.RLock()
	if _, exists := l.m[key]; !exists {
		l.mapLock.RUnlock()
		return nil
	}
	l.mapLock.RUnlock()

	go l.store.Delete(key)
	l.mapLock.Lock()
	delete(l.m, key)
	l.mapLock.Unlock()
	return nil
}

// Update updates an existing entry in the link map. This should only be used to
// update existing entries, as Put is much more efficient for additions.
func (l *LinkMap) Update(key string, target *url.URL) error {
	go l.store.Update(key, target.String())

	l.mapLock.Lock()
	defer l.mapLock.Unlock()

	l.m[key] = target.String()
	return nil
}

func (e *ParseError) Error() string {
	return ""
}
