package links

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/rs/zerolog/log"
)

func TestLinkMap_Delete(t *testing.T) {
	links := NewLinkMap(storage.NONE, "")
	urls, err := testSetup("https://foo.com", "https://bar.com")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup test urls")
	}

	links.Put("foo", urls[0])
	links.Put("bar", urls[1])

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
			links.Delete(tt.key)
			if _, existsActual := links.Get(tt.key); tt.present != existsActual {
				t.Errorf("Expected entry %s to not be present", tt.key)
			}
		})
	}
}

func TestLinkMap_Get(t *testing.T) {
	links := NewLinkMap(storage.NONE, "")
	urls, err := testSetup("https://foo.com", "https://bar.com")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to setup test urls")
	}

	links.Put("foo", urls[0])
	links.Put("bar", urls[1])

	tests := []struct {
		name    string
		key     string
		value   string
		present bool
	}{
		{name: "Read existing entry", key: "foo", value: "https://foo.com", present: true},
		{name: "Read other existing entry", key: "bar", value: "https://bar.com", present: true},
		{name: "Read nonexistent entry", key: "baz", value: "", present: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, exists := links.Get(tt.key)
			if exists != tt.present {
				var expectedLog string
				if tt.present {
					expectedLog = "to be present"
				} else {
					expectedLog = "not to be present"
				}
				t.Errorf("Expected entry %s to %s", tt.key, expectedLog)
			}
			if val != tt.value {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", tt.key, tt.value, val)
			}
		})
	}
}

func TestLinkMap_GetFiltered(t *testing.T) {
	links := NewLinkMap(storage.NONE, "")
	urls, err := testSetup("https://foo.com", "https://bar.com", "https://foobar.com")
	if err != nil {
		t.Errorf("Failed to setup test urls: %v", err)
	}

	links.Put("foo", urls[0])
	links.Put("bar", urls[1])
	links.Put("foobar", urls[2])

	tests := []struct {
		name     string
		keys     []string
		expected map[string]string
	}{
		{
			name:     "Fetches multiple entries",
			keys:     []string{"foo", "bar"},
			expected: map[string]string{"foo": "https://foo.com", "bar": "https://bar.com"},
		},
		{
			name:     "Fetches single entry",
			keys:     []string{"bar"},
			expected: map[string]string{"bar": "https://bar.com"},
		},
		{
			name:     "Does not error with nonexistent entry",
			keys:     []string{"baz"},
			expected: map[string]string{},
		},
		{
			name: "Does not prefix match",
			keys: []string{"foo"},
			expected: map[string]string{
				"foo": "https://foo.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := links.GetFiltered(tt.keys)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", tt.keys, tt.expected, actual)
			}
		})
	}
}

func TestLinkMap_Put(t *testing.T) {
	links := NewLinkMap(storage.NONE, "")
	urls, err := testSetup("https://foo.com", "https://bar.com")
	if err != nil {
		t.Errorf("Failed to setup test urls: %v", err)
	}

	tests := []struct {
		name  string
		key   string
		value *url.URL
	}{
		{name: "put a new entry", key: "foo", value: urls[0]},
		{name: "put an existing entry", key: "foo", value: urls[1]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links.Put(tt.key, tt.value)
			if val, exists := links.Get(tt.key); !exists || val != tt.value.String() {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", tt.key, tt.value, val)
			}
		})
	}
}

func TestLinkMap_Update(t *testing.T) {
	links := NewLinkMap(storage.NONE, "")
	urls, err := testSetup("https://foo.com", "https://bar.com")
	if err != nil {
		t.Errorf("Failed to setup test urls: %v", err)
	}

	tests := []struct {
		name  string
		key   string
		value *url.URL
	}{
		{name: "put a new entry", key: "foo", value: urls[0]},
		{name: "put an existing entry", key: "foo", value: urls[1]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			links.Update(tt.key, tt.value)
			if val, exists := links.Get(tt.key); !exists || val != tt.value.String() {
				t.Errorf("Expected entry %s to contain %s, got %s instead.", tt.key, tt.value, val)
			}
		})
	}
}

func testSetup(urlStrings ...string) ([]*url.URL, error) {
	urls := make([]*url.URL, len(urlStrings))
	for i, urlString := range urlStrings {
		url, err := url.Parse(urlString)
		if err != nil {
			return nil, err
		}
		urls[i] = url
	}
	return urls, nil
}
