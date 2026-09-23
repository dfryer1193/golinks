package links

import (
	"io"
	"net/url"

	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/rs/zerolog/log"
)

type LinkMap interface {
	Get(key string) (string, bool)
	GetAll() map[string]string
	GetAllKeys() []string
	GetFiltered(keys []string) map[string]string

	Put(key string, target *url.URL) error
	Delete(key string) error
	Update(key string, target *url.URL) error
	ReplaceAll(mapReader io.Reader) error
}

type ParseError struct{}

func (e *ParseError) Error() string {
	return ""
}

func buildStorage(persistType storage.StorageType, requestedConfig string) storage.Storage {
	switch persistType {
	case storage.NONE:
		return storage.NewNoneStorage()
	case storage.FILE:
		return storage.NewFileStorage(requestedConfig)
	case storage.SQLITE:
		s, err := storage.NewSQLiteStorage(requestedConfig)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize SQLite storage")
			return storage.NewNoneStorage()
		}
		return s
	case storage.POSTGRES:
		s, err := storage.NewPostgresStorage(requestedConfig)
		if err != nil {
			log.Error().Err(err).Msg("Failed to initialize Postgres storage")
			return storage.NewNoneStorage()
		}
		return s
	default:
		return storage.NewFileStorage("")
	}
}
