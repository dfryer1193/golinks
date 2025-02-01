package storage

import (
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"net/url"
	"os"
	"reflect"
	"sync"
	"testing"
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
	f.Put("baz", "https://baz.com")
	m := f.Read()
	if actual := m["baz"]; actual != "https://baz.com" {
		log.Fatal().Msgf("Expected entry baz to contain https://baz.com, got %s instead.", actual)
	}
	f.Put("baz", "https://abc.com")
	m = f.Read()
	if actual := m["baz"]; actual != "https://abc.com" {
		log.Fatal().Msgf("Expected entry baz to contain https://abc.com, got %s instead.", actual)
	}
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

func TestFileStorage_watchConfig(t *testing.T) {
	type fields struct {
		configPath    string
		fileLock      *sync.RWMutex
		watcher       *fsnotify.Watcher
		reloadChannel chan bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileStorage{
				configPath:    tt.fields.configPath,
				fileLock:      tt.fields.fileLock,
				watcher:       tt.fields.watcher,
				reloadChannel: tt.fields.reloadChannel,
			}
			f.watchConfig()
		})
	}
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
