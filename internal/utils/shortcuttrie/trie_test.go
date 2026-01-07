package shortcuttrie

import (
	"testing"
)

func TestPathsWithoutVariables(t *testing.T) {
	inputPath := "/foo/bar/baz"
	inputTarget := "http://example.com/foo/bar/baz"

	trie := NewShortcutTrie()
	trie.Insert(inputPath, inputTarget)

	retrievedTarget, found := trie.Get("/foo/bar/baz")
	if !found {
		t.Errorf("Expected to find path %s, but did not.", inputPath)
	}

	if retrievedTarget != inputTarget {
		t.Errorf("Expected target %s, got %s instead.", inputTarget, retrievedTarget)
	}
}

func TestPathsWithVariables(t *testing.T) {
	inputPath := "/foo/{var}/baz"
	inputTarget := "http://example.com/foo/{var}/baz"

	trie := NewShortcutTrie()
	trie.Insert(inputPath, inputTarget)

	retrievedTarget, found := trie.Get("/foo/somevalue/baz")
	if !found {
		t.Errorf("Expected to find path %s, but did not.", inputPath)
	}

	expectedTarget := "http://example.com/foo/somevalue/baz"
	if retrievedTarget != expectedTarget {
		t.Errorf("Expected target %s, got %s instead.", expectedTarget, retrievedTarget)
	}
}

func TestPathsWithMultipleVariables(t *testing.T) {
	inputPath := "/{var1}/bar/{var2}"
	inputTarget := "http://example.com/{var1}/bar/{var2}"

	trie := NewShortcutTrie()
	trie.Insert(inputPath, inputTarget)

	retrievedTarget, found := trie.Get("/foo/bar/baz")
	if !found {
		t.Errorf("Expected to find path %s, but did not.", inputPath)
	}

	expectedTarget := "http://example.com/foo/bar/baz"
	if retrievedTarget != expectedTarget {
		t.Errorf("Expected target %s, got %s instead.", expectedTarget, retrievedTarget)
	}
}

func TestNotFoundForNonexistentPaths(t *testing.T) {
	trie := NewShortcutTrie()

	_, found := trie.Get("/nonexistent/path")
	if found {
		t.Errorf("Expected not to find path /nonexistent/path, but it was found.")
	}
}

func TestHandlesOverlappingPaths(t *testing.T) {
	trie := NewShortcutTrie()
	trie.Insert("/foo/bar/qux", "http://example.com/foo/bar/qux")
	trie.Insert("/foo/{var}", "http://example.com/foo/{var}")

	retrievedTarget1, found1 := trie.Get("/foo/bar/qux")
	if !found1 {
		t.Errorf("Expected to find path /foo/bar, but did not.")
	}
	expectedTarget1 := "http://example.com/foo/bar/qux"
	if retrievedTarget1 != expectedTarget1 {
		t.Errorf("Expected target %s, got %s instead.", expectedTarget1, retrievedTarget1)
	}

	retrievedTarget2, found2 := trie.Get("/foo/baz")
	if !found2 {
		t.Errorf("Expected to find path /foo/baz, but did not.")
	}
	expectedTarget2 := "http://example.com/foo/baz"
	if retrievedTarget2 != expectedTarget2 {
		t.Errorf("Expected target %s, got %s instead.", expectedTarget2, retrievedTarget2)
	}
}

func TestMidPathVariables(t *testing.T) {
	inputPath := "/foo/{var}/baz/qux"
	inputTarget := "http://example.com/foo/{var}/baz/qux"

	trie := NewShortcutTrie()
	trie.Insert(inputPath, inputTarget)

	retrievedTarget, found := trie.Get("/foo/somevalue/baz/qux")
	if !found {
		t.Errorf("Expected to find path %s, but did not.", inputPath)
	}

	expectedTarget := "http://example.com/foo/somevalue/baz/qux"
	if retrievedTarget != expectedTarget {
		t.Errorf("Expected target %s, got %s instead.", expectedTarget, retrievedTarget)
	}
}
