package storage

import (
	"database/sql"
	"io"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			key TEXT PRIMARY KEY,
			target TEXT NOT NULL
		)
	`); err != nil {
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

func (s *SQLiteStorage) Read() (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, target FROM links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make(map[string]string)
	for rows.Next() {
		var key, target string
		if err := rows.Scan(&key, &target); err != nil {
			return nil, err
		}
		links[key] = target
	}

	return links, nil
}

func (s *SQLiteStorage) Put(key string, target string) {
	if _, err := s.db.Exec("INSERT INTO links (key, target) VALUES (?, ?)", key, target); err != nil {
		log.Error().Err(err).Msg("failed to insert link")
	}
}

func (s *SQLiteStorage) Delete(key string) {
	if _, err := s.db.Exec("DELETE FROM links WHERE key = ?", key); err != nil {
		log.Error().Err(err).Msg("failed to delete link")
	}
}

func (s *SQLiteStorage) Update(key string, target string) {
	if _, err := s.db.Exec("UPDATE links SET target = ? WHERE key = ?", target, key); err != nil {
		log.Error().Err(err).Msg("failed to update link")
	}
}

func (s *SQLiteStorage) GetReloadChannel() <-chan bool {
	// For SQLite, we don't need to reload the config from a file, so we can return a nil channel.
	return nil
}

func (s *SQLiteStorage) ReplaceConfig(reader io.Reader) (map[string]string, error) {
	// This is not applicable to SQLite storage, so we'll return an empty map and no error.
	return make(map[string]string), nil
}
