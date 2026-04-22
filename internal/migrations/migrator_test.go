package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestMigratorSQLite(t *testing.T) {
	dbPath := "test_migrations.db"
	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create migrator
	migrator, err := NewMigrator(db, "sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	// Check that migrations were loaded
	if len(migrator.migrations) == 0 {
		t.Error("No migrations loaded")
	}

	// Run migrations
	if err := migrator.Migrate(); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify links table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='links'").Scan(&tableName)
	if err != nil {
		t.Fatalf("Links table not found: %v", err)
	}
	if tableName != "links" {
		t.Errorf("Expected table name 'links', got '%s'", tableName)
	}

	// Verify schema_migrations table exists
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	if err != nil {
		t.Fatalf("schema_migrations table not found: %v", err)
	}

	// Verify migration was recorded
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count migrations: %v", err)
	}
	if count == 0 {
		t.Error("No migrations recorded in schema_migrations table")
	}

	// Test idempotency - running migrations again should succeed
	if err := migrator.Migrate(); err != nil {
		t.Fatalf("Second migration run failed: %v", err)
	}

	// Count should still be the same
	var newCount int
	err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&newCount)
	if err != nil {
		t.Fatalf("Failed to count migrations after second run: %v", err)
	}
	if newCount != count {
		t.Errorf("Migration count changed from %d to %d (should be idempotent)", count, newCount)
	}
}

func TestMigratorInvalidDBType(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = NewMigrator(db, "mysql")
	if err == nil {
		t.Error("Expected error for unsupported database type, got nil")
	}
}

func TestMigratorLoadMigrations(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	migrator, err := NewMigrator(db, "sqlite")
	if err != nil {
		t.Fatalf("Failed to create migrator: %v", err)
	}

	// Verify migrations are sorted by version
	for i := 1; i < len(migrator.migrations); i++ {
		if migrator.migrations[i-1].Version >= migrator.migrations[i].Version {
			t.Errorf("Migrations not sorted: %s >= %s",
				migrator.migrations[i-1].Version,
				migrator.migrations[i].Version)
		}
	}

	// Verify each migration has required fields
	for _, m := range migrator.migrations {
		if m.Version == "" {
			t.Error("Migration has empty version")
		}
		if m.Name == "" {
			t.Error("Migration has empty name")
		}
		if m.SQL == "" {
			t.Error("Migration has empty SQL")
		}
	}
}
