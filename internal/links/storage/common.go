package storage

type Storage interface {
	Read() map[string]string
	Put(key string, target string)
	Delete(key string)
	Update(key string, target string)
	GetReloadChannel() <-chan bool
}

type StorageType int

const (
	NONE StorageType = iota
	FILE
)

type ParseError struct{}

func (e *ParseError) Error() string {
	return "Failed to parse config"
}
