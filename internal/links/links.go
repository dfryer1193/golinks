package links

import (
	"io"
	"net/url"

	"github.com/dfryer1193/golinks/internal/links/storage"
)

type LinkMap interface {
	Get(key string) (string, bool)
	GetAll() map[string]string
	GetAllKeys() []string
	GetFiltered(keys []string) map[string]string

	// TODO: Make target a string so we can support parameterized links
	Put(key string, target *url.URL) error
	Delete(key string) error
	// TODO: Make target a string so we can support parameterized links
	Update(key string, target *url.URL) error
	ReplaceAll(mapReader io.Reader) error
	Close() error
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
			panic(err)
		}
		return s
	case storage.POSTGRES:
		s, err := storage.NewPostgresStorage(requestedConfig)
		if err != nil {
			panic(err)
		}
		return s
	default:
		return storage.NewFileStorage("")
	}
}
