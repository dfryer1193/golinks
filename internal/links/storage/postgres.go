package storage

import (
	"database/sql"
	"fmt"
	"io"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var _ Storage = (*PostgresStorage)(nil)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(connString string) (*PostgresStorage, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for PostgreSQL
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}

	// Validate schema exists
	if err := storage.validateSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

func (s *PostgresStorage) validateSchema() error {
	var exists bool
	err := s.db.QueryRow(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'links')",
	).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check schema: %w", err)
	}
	if !exists {
		return fmt.Errorf("schema not initialized: please run 'golinks migrate -storage POSTGRES' first")
	}
	return nil
}

func (s *PostgresStorage) Read() (map[string]string, error) {
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

	return links, rows.Err()
}

func (s *PostgresStorage) Get(key string) (string, bool) {
	var target string
	err := s.db.QueryRow("SELECT target FROM links WHERE key = $1", key).Scan(&target)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		log.Error().Err(err).Msg("failed to get link")
		return "", false
	}
	return target, true
}

func (s *PostgresStorage) Put(key string, target string) error {
	_, err := s.db.Exec(
		"INSERT INTO links (key, target) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET target = EXCLUDED.target",
		key, target,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert or replace link")
	}
	return err
}

func (s *PostgresStorage) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM links WHERE key = $1", key)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete link")
	}
	return err
}

func (s *PostgresStorage) Update(key string, target string) error {
	_, err := s.db.Exec("UPDATE links SET target = $1 WHERE key = $2", target, key)
	if err != nil {
		log.Error().Err(err).Msg("failed to update link")
	}
	return err
}

func (s *PostgresStorage) GetReloadChannel() <-chan bool {
	return nil
}

func (s *PostgresStorage) ReplaceConfig(reader io.Reader) (map[string]string, error) {
	newLinks, err := parseLinksFile(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse links file: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM links"); err != nil {
		return nil, fmt.Errorf("failed to clear existing links: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO links (key, target) VALUES ($1, $2)")
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

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
