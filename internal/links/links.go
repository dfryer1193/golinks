package links

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type ParseError struct{}

// LinkMap houses the map of redirects, and keeps track of the backing file for
// maintaining the map across restarts. It also handles thread safety.
type LinkMap struct {
	configPath string
	m          map[string]string
	mapLock    *sync.RWMutex
	fileLock   *sync.RWMutex
}

// NewLinkMap generates a new LinkMap object, with the requested config if it
// exists. If the requested config does not exist, it falls back to the default
// locations. If the default locations do not exist, the program will exit with
// an error.
func NewLinkMap(requestedConfig string) *LinkMap {
	path, config := findConfig(requestedConfig)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating file watcher")
	}

	linkMap := LinkMap{
		configPath: path,
		m:          parseConfig(config),
		mapLock:    &sync.RWMutex{},
		fileLock:   &sync.RWMutex{},
	}

	go linkMap.watchConfig(watcher)

	return &linkMap
}

func getHomeDir() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get current user")
	}

	return currentUser.HomeDir
}

func findConfig(requestedConfig string) (string, *os.File) {
	homedir := getHomeDir()
	configs := []string{
		requestedConfig,
		"./links",
		homedir + "/.config/golinks/links",
		"/etc/golinks/links",
	}
	errs := []error{}

	for _, config := range configs {
		if config == "" {
			continue
		}
		file, err := openFile(config)
		if err == nil { // Note the deviation from the standard err != nil
			log.Info().Msg("Using config file " + config)
			return config, file
		}
		errs = append(errs, err)
	}

	for _, err := range errs {
		log.Err(err)
	}
	log.Fatal().Msg("Failed to find any config")
	return "", nil
}

func openFile(path string) (*os.File, error) {
	fpath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func parseLine(line string, lineNum int) (string, *url.URL) {
	parts := strings.FieldsFunc(strings.TrimSpace(line), func(c rune) bool { return unicode.IsSpace(c) })
	if len(parts) != 2 {
		if len(parts) == 0 {
			return "", nil
		}
		err := &ParseError{}
		log.Fatal().Err(err).Int("line", lineNum).Msg("Malformed config. Each non-empty line must have exactly two entries.")
		panic(1)
	}

	target, err := url.Parse(parts[1])
	if err != nil {
		log.Err(err).Int("line", lineNum).Str("url", parts[1]).Msg("Malformed config. Invalid url")
	}

	return parts[0], target
}

func parseConfig(filePtr *os.File) map[string]string {
	linkMap := make(map[string]string)
	defer filePtr.Close()

	scanner := bufio.NewScanner(filePtr)

	lineNum := 0
	for scanner.Scan() {
		txt := scanner.Text()
		lineNum++
		key, target := parseLine(txt, lineNum)
		if key == "" && target == nil {
			continue
		}
		linkMap[key] = target.String()
	}

	return linkMap
}

func (l *LinkMap) watchConfig(watcher *fsnotify.Watcher) {
	dir := filepath.Dir(l.configPath)
	name := filepath.Base(l.configPath)
	err := watcher.Add(dir)
	if err != nil {
		log.Err(err).Msg("Failed to add watcher on config dir. Config will not live reload")
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Debug().Msg("File watch event received!")
			if filepath.Base(event.Name) == name && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				log.Debug().Msg("Config file updated, reloading...")
				l.update()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			if err != nil {
				log.Err(err).Msg("File watch error received")
			}
			fmt.Println(err)
		}
	}
}

func (l *LinkMap) update() {
	file, err := os.OpenFile(l.configPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Str("file", l.configPath).Msg("Cound not open config for reading")
	}
	defer file.Close()

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	log.Debug().Str("file", l.configPath).Msg("Reading config file")
	l.m = parseConfig(file)
}

// Get returns the url and state of existence for a single key.
func (l *LinkMap) Get(key string) (string, bool) {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()
	target, exists := l.m[key]
	return target, exists
}

// GetAll returns a map containing all of the entries from the current LinkMap object
func (l *LinkMap) GetAll() map[string]string {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()

	return l.m
}

func (l *LinkMap) GetAllKeys() []string {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()

	keys := make([]string, len(l.m))
	i := 0
	for key := range l.m {
		keys[i] = key
		i++
	}

	return keys
}

func (l *LinkMap) GetFiltered(keys []string) map[string]string {
	filteredMap := make(map[string]string, len(keys))

	for _, key := range keys {
		filteredMap[key] = l.m[key]
	}

	return filteredMap
}

// Put appends a new entry to the link map. If the entry already exists, it will
// be duplicated in the backing file, and the value in the live map will be
// replaced.
func (l *LinkMap) Put(key string, target *url.URL) error {
	l.fileLock.Lock()
	defer l.fileLock.Unlock()

	file, err := os.OpenFile(l.configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(key + " " + target.String() + "\n"); err != nil {
		return err
	}

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	l.m[key] = target.String()

	return nil
}

// Delete removes an entry from the link map. If the key is not present in the
// map, this is a no-op
func (l *LinkMap) Delete(key string) error {
	// Can skip filesystem-intensive writes if the entry already doesn't exist
	l.mapLock.RLock()
	if _, exists := l.m[key]; !exists {
		return nil
	}
	l.mapLock.RUnlock()

	err := l.updateEntry(key, nil)
	if err != nil {
		return err
	}

	return nil
}

// Update updates an existing entry in the link map. This should only be used to
// update existing entries, as Put is much more efficient for additions.
func (l *LinkMap) Update(key string, target *url.URL) error {
	err := l.updateEntry(key, target)
	if err != nil {
		return err
	}

	return nil
}

func (l *LinkMap) updateEntry(key string, target *url.URL) error {
	l.fileLock.Lock()
	curFile, err := os.OpenFile(l.configPath, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}

	newFile, err := os.OpenFile(l.getScratchConfigFilepath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(curFile)
	for scanner.Scan() {
		txt := scanner.Text()
		if strings.HasPrefix(txt, key+" ") {
			if target == nil {
				continue
			}
			if _, err := newFile.WriteString(key + " " + target.String() + "\n"); err != nil {
				return err
			}
			continue
		}
		_, err := newFile.WriteString(txt + "\n")
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	curFile.Close()
	newFile.Close()
	l.fileLock.Unlock()

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	if target == nil {
		delete(l.m, key)
	} else {
		l.m[key] = target.String()
	}

	err = l.replaceConfigInPlace()
	if err != nil {
		return err
	}

	return nil
}

func (l *LinkMap) getScratchConfigFilepath() string {
	return l.configPath + "~"
}

func (l *LinkMap) getBackupConfigFilepath() string {
	return l.configPath + ".bak"
}

func (l *LinkMap) replaceConfigInPlace() error {
	l.fileLock.Lock()
	defer l.fileLock.Unlock()
	err := os.Rename(l.configPath, l.getBackupConfigFilepath())
	if err != nil {
		return err
	}
	err = os.Rename(l.getScratchConfigFilepath(), l.configPath)
	if err != nil {
		return err
	}
	err = os.Remove(l.getBackupConfigFilepath())
	if err != nil {
		return err
	}
	return nil
}

func (e *ParseError) Error() string {
	return ""
}
