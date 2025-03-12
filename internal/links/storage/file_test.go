package storage

import (
	"github.com/rs/zerolog/log"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
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
		log.Fatal().Err(err).Msg("Failed to create test file")
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
			entries := f.Read()
			_, exists := entries[tt.key]
			if exists != tt.present {
				log.Fatal().Msgf("Expected entry %s to not be present", tt.key)
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
			entries := f.Read()
			actual := entries[tt.key]
			if actual != tt.target {
				log.Fatal().Msgf("Expected entry %s to contain %s, got %s instead.", entries[tt.key], tt.target, actual)
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

			actual := f.Read()
			if actual[tt.key] != tt.target {
				log.Fatal().Msgf("Expected entry %s to contain %s, got %s instead.", actual, tt.target, actual)
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
			entries := f.Read()
			actual := entries[tt.key]
			if actual != tt.target {
				log.Fatal().Msgf("Expected entry %s to contain %s, got %s instead.", entries[tt.key], tt.target, actual)
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
					log.Fatal().Msg("Got unexpected reload signal")
				}
			case <-time.After(time.Millisecond * 1000):
				if tt.expectReload {
					log.Fatal().Msg("Did not receive expected reload signal")
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
		name  string
		args  args
		want  string
		want1 *url.URL
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseLine(tt.args.line, tt.args.lineNum)
			if got != tt.want {
				t.Errorf("parseLine() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseLine() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
