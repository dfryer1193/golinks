package storage

import (
	"github.com/rs/zerolog/log"
	"strings"
)

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

func (st StorageType) String() string {
	return [...]string{"NONE", "FILE"}[st]
}

func FromString(s string) StorageType {
	sanitized := strings.ToUpper(s)
	switch sanitized {
	case "NONE":
		return NONE
	case "FILE":
		return FILE
	default:
		log.Fatal().Str("requestedStorageType", s).Msg("Storage type not recognized")
	}
	return FILE
}

type ParseError struct{}

func (e *ParseError) Error() string {
	return "Failed to parse config"
}
