package storage

import (
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

const TEST_DIR = "./test"
const TEST_FILE = "test.links"

func createTestFile() {
	fileInfo, err := os.Stat(TEST_DIR)
	if os.IsNotExist(err) {
		err := os.Mkdir(TEST_DIR, 0777)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create test dir")
		}
	}

	if !fileInfo.IsDir() {
		log.Fatal().Msg("Test dir is not a directory")
	}

	file, err := os.Create(TEST_DIR + "/" + TEST_FILE)
	defer file.Close()
	if err != nil {
		if !os.IsExist(err) {
			log.Fatal().Err(err).Msg("Failed to create test file")
		} else {
			cleanup()
		}
	}

	file.WriteString("foo https://test.com\n")
	file.WriteString("bar https://example.app\n")
}

func cleanup() {
	err := os.Remove(TEST_DIR + "/" + TEST_FILE)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to remove test file")
	}
}

func TestFileStorage_Delete(t *testing.T) {
	createTestFile()
	f := NewFileStorage(TEST_DIR + "/" + TEST_FILE)
	tests := []struct {
		name    string
		key     string
		present bool
	}{
		{name: "delete existing entry", key: "foo", present: false},
		{name: "delete nonexistent entry", key: "foo", present: false},
		{name: "delete last entry", key: "bar", present: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.Delete(tt.key)
			entries, err := f.Read()
			if err != nil {
				t.Errorf("failed to read test entries: %v", err)
			}
			_, exists := entries[tt.key]
			if exists != tt.present {
				t.Errorf("Expected entry %s to not be present", tt.key)
			}
		})
	}
	cleanup()
}

func TestFileStorage_Put(t *testing.T) {
	createTestFile()
	f := NewFileStorage(TEST_DIR + "/" + TEST_FILE)
	tests := []struct {
		name   string
		key    string
		target string
	}{
		{name: "put existing entry", key: "foo", target: "https://abc.com"},
		{name: "add a new entry", key: "baz", target: "https://baz.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.Put(tt.key, tt.target)
			entries, err := f.Read()
			if err != nil {
				t.Errorf("failed to read test entries: %v", err)
			}
			actual := entries[tt.key]
			if actual != tt.target {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", entries[tt.key], tt.target, actual)
			}
		})
	}
	cleanup()
}

func TestFileStorage_Read(t *testing.T) {
	createTestFile()
	f := NewFileStorage(TEST_DIR + "/" + TEST_FILE)
	tests := []struct {
		name      string
		operation string
		key       string
		target    string
	}{
		{name: "Reads after new entry", operation: "put", key: "baz", target: "https://baz.com"},
		{name: "Reads last target when duplicated", operation: "put", key: "baz", target: "https://abc.com"},
		{name: "Reads after updating target", operation: "update", key: "foo", target: "https://foo.com"},
		{name: "Reads after deleting target", operation: "delete", key: "foo", target: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operation {
			case "put":
				f.Put(tt.key, tt.target)
			case "update":
				f.Update(tt.key, tt.target)
			case "delete":
				f.Delete(tt.key)
			}

			actual, err := f.Read()
			if err != nil {
				t.Errorf("failed to read test entries: %v", err)
			}
			if actual[tt.key] != tt.target {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", actual[tt.key], tt.target, actual[tt.key])
			}
		})
	}
	cleanup()
}

func TestFileStorage_Update(t *testing.T) {
	createTestFile()
	f := NewFileStorage(TEST_DIR + "/" + TEST_FILE)
	tests := []struct {
		name   string
		key    string
		target string
	}{
		{name: "update existing entry", key: "foo", target: "https://abc.com"},
		{name: "update a new entry", key: "baz", target: "https://baz.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f.Put(tt.key, tt.target)
			entries, err := f.Read()
			if err != nil {
				t.Errorf("failed to read test entries: %v", err)
			}
			actual := entries[tt.key]
			if actual != tt.target {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", entries[tt.key], tt.target, actual)
			}
		})
	}
	cleanup()
}

func TestFileStorage_ReloadSignaling(t *testing.T) {
	createTestFile()
	f := NewFileStorage(TEST_DIR + "/" + TEST_FILE)
	reloadChannel := f.GetReloadChannel()
	tests := []struct {
		name         string
		operation    string
		key          string
		target       string
		expectReload bool
	}{
		{name: "Sends reload signal after new entry", operation: "put", key: "baz", target: "https://baz.com", expectReload: true},
		{name: "Does not send reload signal after read", operation: "read", expectReload: false},
		{name: "Sends reload signal after updating target", operation: "update", key: "foo", target: "https://foo.com", expectReload: true},
		{name: "Sends reload signal when deleting target", operation: "delete", key: "foo", expectReload: true},
		{name: "Does not send reload signal when delete does not change the file", operation: "delete", key: "foo", expectReload: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.operation {
			case "put":
				f.Put(tt.key, tt.target)
			case "update":
				f.Update(tt.key, tt.target)
			case "delete":
				f.Delete(tt.key)
			case "read":
				f.Read()
			}

			select {
			case reload := <-reloadChannel:
				if reload != tt.expectReload {
					t.Errorf("Expected reload signal to be %v, got %v instead.", tt.expectReload, reload)
				}
			case <-time.After(time.Millisecond * 1000):
				if tt.expectReload {
					t.Errorf("Expected reload signal to be %v, got none instead.", tt.expectReload)
				}
			}
		})
	}
	cleanup()
}

func Test_parseLine(t *testing.T) {
	type args struct {
		line    string
		lineNum int
	}
	tests := []struct {
		name       string
		args       args
		wantKey    string
		wantTarget *url.URL
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, target, err := parseLine(tt.args.line, tt.args.lineNum)
			if err != nil {
				t.Errorf("parseLine() error = %v", err)
				return
			}
			if key != tt.wantKey {
				t.Errorf("parseLine() got = %v, want %v", key, tt.wantKey)
			}
			if !reflect.DeepEqual(target, tt.wantTarget) {
				t.Errorf("parseLine() got1 = %v, want %v", target, tt.wantTarget)
			}
		})
	}
}
