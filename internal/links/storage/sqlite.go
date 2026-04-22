package storage

import (
	"database/sql"
	"fmt"
	"io"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

var _ Storage = (*SQLiteStorage)(nil)

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

func (s *SQLiteStorage) Get(key string) (string, bool) {
	var target string
	err := s.db.QueryRow("SELECT target FROM links WHERE key = ?", key).Scan(&target)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		log.Error().Err(err).Msg("failed to get link")
		return "", false
	}

	return target, true
}

func (s *SQLiteStorage) Put(key string, target string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO links (key, target) VALUES (?, ?)", key, target)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert or replace link")
	}
	return err
}

func (s *SQLiteStorage) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM links WHERE key = ?", key)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete link")
	}
	return err
}

func (s *SQLiteStorage) Update(key string, target string) error {
	_, err := s.db.Exec("UPDATE links SET target = ? WHERE key = ?", target, key)
	if err != nil {
		log.Error().Err(err).Msg("failed to update link")
	}
	return err
}

func (s *SQLiteStorage) GetReloadChannel() <-chan bool {
	// For SQLite, we don't need to reload the config from a file, so we can return a nil channel.
	return nil
}

func (s *SQLiteStorage) ReplaceConfig(reader io.Reader) (map[string]string, error) {
	// Parse the input reader to get the new links
	newLinks, err := parseLinksFile(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse links file: %w", err)
	}

	// Clear the existing links table
	if _, err := s.db.Exec("DELETE FROM links"); err != nil {
		return nil, fmt.Errorf("failed to clear existing links: %w", err)
	}

	// Insert all new links into the database
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO links (key, target) VALUES (?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for key, target := range newLinks {
		if _, err := stmt.Exec(key, target); err != nil {
			return nil, fmt.Errorf("failed to insert link %s: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newLinks, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
