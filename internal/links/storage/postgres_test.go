package storage

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func TestPostgreSQLStorage(t *testing.T) {
	// Skip if DATABASE_URL not set (for CI/CD environments without PostgreSQL)
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping PostgreSQL tests")
	}

	// Create schema manually since we removed auto-schema creation
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	// Clean up test table if exists
	_, err = db.Exec(`DROP TABLE IF EXISTS links`)
	if err != nil {
		t.Fatalf("DROP TABLE error = %v", err)
	}

	// Create schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			key TEXT PRIMARY KEY,
			target TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	storage, err := NewPostgresStorage(connStr)
	if err != nil {
		t.Fatalf("NewPostgresStorage() error = %v", err)
	}
	defer storage.Close()

	// Test Put
	if err := storage.Put("test", "https://example.com"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}

	// Test Read
	links, err := storage.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(links) != 1 {
		t.Errorf("Read() got %v links, want 1", len(links))
	}
	if links["test"] != "https://example.com" {
		t.Errorf("Read() got %v, want https://example.com", links["test"])
	}

	// Test Get
	target, exists := storage.Get("test")
	if !exists {
		t.Errorf("Get() key not found")
	}
	if target != "https://example.com" {
		t.Errorf("Get() got %v, want https://example.com", target)
	}

	// Test Update
	if err := storage.Update("test", "https://example.org"); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	links, err = storage.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if links["test"] != "https://example.org" {
		t.Errorf("Read() got %v, want https://example.org", links["test"])
	}

	// Test Delete
	if err := storage.Delete("test"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	links, err = storage.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(links) != 0 {
		t.Errorf("Read() got %v links, want 0", len(links))
	}

	// Clean up test table
	_, err = db.Exec(`DROP TABLE IF EXISTS links`)
	if err != nil {
		t.Fatalf("DROP TABLE cleanup error = %v", err)
	}
}
