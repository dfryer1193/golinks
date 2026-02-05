package storage

import (
	"os"
	"testing"
)

func TestSQLiteStorage(t *testing.T) {
	dbPath := "test.db"
	defer os.Remove(dbPath)

	storage, err := NewSQLiteStorage(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteStorage() error = %v", err)
	}

	// Test Put
	storage.Put("test", "https://example.com")

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
	storage.Update("test", "https://example.org")
	links, err = storage.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if links["test"] != "https://example.org" {
		t.Errorf("Read() got %v, want https://example.org", links["test"])
	}

	// Test Delete
	storage.Delete("test")
	links, err = storage.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(links) != 0 {
		t.Errorf("Read() got %v links, want 0", len(links))
	}
}