package links

import (
	"io"
	"net/url"

	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/rs/zerolog/log"
)

var _ LinkMap = (*BaseLinkMap)(nil)

type BaseLinkMap struct {
	store storage.Storage
}

func NewBaseLinkMap(persistType storage.StorageType, requestedConfig string) *BaseLinkMap {
	if persistType == storage.NONE {
		log.Warn().Msg("Using a 'None' storage type with a base link map will result in a link map that cannot hold any entries.")
	}

	store := buildStorage(persistType, requestedConfig)

	return &BaseLinkMap{store: store}
}

func (l *BaseLinkMap) Get(key string) (string, bool) {
	return l.store.Get(key)
}

func (l *BaseLinkMap) GetAll() map[string]string {
	links, err := l.store.Read()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read links from data store")
		return make(map[string]string)
	}

	return links
}

func (l *BaseLinkMap) GetAllKeys() []string {
	links, err := l.store.Read()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read links from data store")
		return make([]string, 0)
	}

	keys := make([]string, 0, len(links))
	for key := range links {
		keys = append(keys, key)
	}

	return keys
}

func (l *BaseLinkMap) GetFiltered(keys []string) map[string]string {
	links, err := l.store.Read()
	if err != nil {
		log.Error().Err(err).Msg("Failed to read links from data store")
		return make(map[string]string)
	}

	filtered := make(map[string]string)
	for _, k := range keys {
		val, exists := links[k]
		if !exists {
			continue
		}

		filtered[k] = val
	}

	return filtered
}

func (l *BaseLinkMap) Put(key string, target *url.URL) error {
	l.store.Put(key, target.String())

	return nil
}

func (l *BaseLinkMap) Delete(key string) error {
	l.store.Delete(key)

	return nil
}

func (l *BaseLinkMap) Update(key string, target *url.URL) error {
	l.store.Update(key, target.String())

	return nil
}

func (l *BaseLinkMap) ReplaceAll(reader io.Reader) error {
	_, err := l.store.ReplaceConfig(reader)
	return err
}
