package storage

import (
	"database/sql"
	"fmt"
	"io"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var _ Storage = (*PostgresStorage)(nil)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connStr string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS links (key TEXT PRIMARY KEY, target TEXT NOT NULL)`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &PostgresStorage{db: db}, nil
}

func (s *PostgresStorage) Read() (map[string]string, error) {
	rows, err := s.db.Query("SELECT key, target FROM links")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	links := make(map[string]string)
	for rows.Next() {
		var k, t string
		if err := rows.Scan(&k, &t); err != nil {
			return nil, err
		}
		links[k] = t
	}
	return links, rows.Err()
}

func (s *PostgresStorage) Get(key string) (string, bool) {
	var target string
	err := s.db.QueryRow("SELECT target FROM links WHERE key = $1", key).Scan(&target)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		log.Error().Err(err).Msg("postgres get failed")
		return "", false
	}
	return target, true
}

func (s *PostgresStorage) Put(key string, target string) error {
	_, err := s.db.Exec("INSERT INTO links (key, target) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET target = $2", key, target)
	return err
}

func (s *PostgresStorage) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM links WHERE key = $1", key)
	return err
}

func (s *PostgresStorage) Update(key string, target string) error {
	_, err := s.db.Exec("UPDATE links SET target = $1 WHERE key = $2", target, key)
	return err
}

func (s *PostgresStorage) GetReloadChannel() <-chan bool { return nil }

func (s *PostgresStorage) ReplaceConfig(reader io.Reader) (map[string]string, error) {
	newLinks, err := parseLinksFile(reader)
	if err != nil {
		return nil, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec("DELETE FROM links")
	if err != nil {
		return nil, err
	}
	stmt, err := tx.Prepare("INSERT INTO links (key, target) VALUES ($1, $2)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for k, v := range newLinks {
		if _, err := stmt.Exec(k, v); err != nil {
			return nil, err
		}
	}
	return newLinks, tx.Commit()
}
