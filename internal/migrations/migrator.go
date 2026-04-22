package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dbmigrations "github.com/dfryer1193/golinks/db/migrations"
	"github.com/rs/zerolog/log"
)

type Migration struct {
	Version string
	Name    string
	SQL     string
}

type Migrator struct {
	db         *sql.DB
	dbType     string
	migrations []Migration
}

// NewMigrator creates a new migration manager for the specified database type
func NewMigrator(db *sql.DB, dbType string) (*Migrator, error) {
	if dbType != "sqlite" && dbType != "postgres" {
		return nil, fmt.Errorf("unsupported database type: %s (supported: sqlite, postgres)", dbType)
	}

	m := &Migrator{
		db:     db,
		dbType: dbType,
	}

	if err := m.loadMigrations(); err != nil {
		return nil, fmt.Errorf("failed to load migrations: %w", err)
	}

	return m, nil
}

// loadMigrations reads migration files from embedded filesystem
func (m *Migrator) loadMigrations() error {
	var fs embed.FS
	var dir string

	switch m.dbType {
	case "sqlite":
		fs = dbmigrations.SqliteMigrations
		dir = "sqlite"
	case "postgres":
		fs = dbmigrations.PostgresMigrations
		dir = "postgres"
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		content, err := fs.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", entry.Name(), err)
		}

		// Parse filename: 20260421000000_initial_schema.sql
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) != 2 {
			log.Warn().Str("file", entry.Name()).Msg("Skipping migration file with invalid name format")
			continue
		}

		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		m.migrations = append(m.migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// ensureMigrationsTable creates the schema_migrations table if it doesn't exist
func (m *Migrator) ensureMigrationsTable() error {
	var createSQL string
	if m.dbType == "sqlite" {
		createSQL = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version TEXT PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	} else {
		createSQL = `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version TEXT PRIMARY KEY,
				applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`
	}

	_, err := m.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	return nil
}

// getAppliedMigrations returns a set of applied migration versions
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	if err := m.ensureMigrationsTable(); err != nil {
		return nil, err
	}

	rows, err := m.db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate() error {
	// Acquire migration lock to prevent concurrent runs
	unlock, err := m.acquireMigrationLock()
	if err != nil {
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	defer unlock()

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	pendingCount := 0
	for _, migration := range m.migrations {
		if applied[migration.Version] {
			continue
		}
		pendingCount++
	}

	if pendingCount == 0 {
		log.Info().Msg("No pending migrations")
		return nil
	}

	log.Info().Int("count", pendingCount).Msg("Running pending migrations")

	for _, migration := range m.migrations {
		if applied[migration.Version] {
			log.Debug().
				Str("version", migration.Version).
				Str("name", migration.Name).
				Msg("Skipping already applied migration")
			continue
		}

		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		log.Info().
			Str("version", migration.Version).
			Str("name", migration.Name).
			Msg("Migration applied successfully")
	}

	log.Info().Msg("All migrations completed successfully")
	return nil
}

// applyMigration applies a single migration within a transaction
func (m *Migrator) applyMigration(migration Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied (idempotent insert)
	var insertSQL string
	if m.dbType == "postgres" {
		// PostgreSQL: INSERT ... ON CONFLICT DO NOTHING
		insertSQL = "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING"
	} else {
		// SQLite: INSERT OR IGNORE
		insertSQL = "INSERT OR IGNORE INTO schema_migrations (version, applied_at) VALUES (?, ?)"
	}

	_, err = tx.Exec(insertSQL, migration.Version, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// acquireMigrationLock acquires a database-level lock for migrations
// Returns an unlock function that must be called to release the lock
func (m *Migrator) acquireMigrationLock() (func() error, error) {
	if m.dbType == "postgres" {
		// PostgreSQL: Use advisory lock
		// Lock ID: 123456789 (arbitrary but consistent for this application)
		const lockID = 123456789
		
		_, err := m.db.Exec("SELECT pg_advisory_lock($1)", lockID)
		if err != nil {
			return nil, fmt.Errorf("failed to acquire advisory lock: %w", err)
		}
		
		unlock := func() error {
			_, err := m.db.Exec("SELECT pg_advisory_unlock($1)", lockID)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to release advisory lock")
				return err
			}
			return nil
		}
		
		log.Debug().Msg("Acquired PostgreSQL advisory lock for migrations")
		return unlock, nil
	} else {
		// SQLite: Use application-level locking via a lock table
		// Create lock table if it doesn't exist
		_, err := m.db.Exec(`
			CREATE TABLE IF NOT EXISTS migration_lock (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				locked_at TIMESTAMP
			)
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to create lock table: %w", err)
		}
		
		// Try to acquire lock with IMMEDIATE transaction
		// This prevents other processes from running migrations concurrently
		// Use a short retry loop to handle contention
		maxRetries := 30
		for i := 0; i < maxRetries; i++ {
			tx, err := m.db.Begin()
			if err != nil {
				return nil, fmt.Errorf("failed to begin transaction: %w", err)
			}
			
			// Try to insert lock row (only one can succeed due to PRIMARY KEY)
			result, err := tx.Exec("INSERT OR IGNORE INTO migration_lock (id, locked_at) VALUES (1, ?)", time.Now())
			if err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to acquire lock: %w", err)
			}
			
			rowsAffected, _ := result.RowsAffected()
			if rowsAffected > 0 {
				// We got the lock, commit to hold it
				if err := tx.Commit(); err != nil {
					return nil, fmt.Errorf("failed to commit lock transaction: %w", err)
				}
				
				unlock := func() error {
					_, err := m.db.Exec("DELETE FROM migration_lock WHERE id = 1")
					if err != nil {
						log.Warn().Err(err).Msg("Failed to release migration lock")
						return err
					}
					return nil
				}
				
				log.Debug().Msg("Acquired SQLite migration lock")
				return unlock, nil
			}
			
			// Lock is held by another process, rollback and retry
			tx.Rollback()
			time.Sleep(100 * time.Millisecond)
		}
		
		return nil, fmt.Errorf("failed to acquire migration lock after %d retries (another process may be running migrations)", maxRetries)
	}
}

// Status returns the current migration status
func (m *Migrator) Status() error {
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	fmt.Printf("\nMigration Status (%s):\n", m.dbType)
	fmt.Println(strings.Repeat("-", 60))

	if len(m.migrations) == 0 {
		fmt.Println("No migrations found")
		return nil
	}

	for _, migration := range m.migrations {
		status := "PENDING"
		if applied[migration.Version] {
			status = "APPLIED"
		}
		fmt.Printf("%-8s  %s  %s\n", status, migration.Version, migration.Name)
	}

	fmt.Println()
	return nil
}
