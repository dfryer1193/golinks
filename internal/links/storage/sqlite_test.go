package storage

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
)

func TestSQLiteStorage(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	// Create schema manually since we removed auto-schema creation
	// Use proper DSN with foreign keys enabled
	connStr := fmt.Sprintf("%s?_foreign_keys=1", dbPath)
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS links (
			key TEXT PRIMARY KEY,
			target TEXT NOT NULL
		)
	`)
	db.Close()
	if err != nil {
		t.Fatalf("CREATE TABLE error = %v", err)
	}

	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStorage() error = %v", err)
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
}