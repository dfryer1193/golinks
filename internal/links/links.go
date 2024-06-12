package links

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"unicode"

	"github.com/fsnotify/fsnotify"
)

type LinkMap struct {
	configPath string
	m          map[string]url.URL
	mapLock    *sync.RWMutex
	fileLock   *sync.RWMutex
}

func NewLinkMap(requestedConfig string) *LinkMap {
	path, config := findConfig(requestedConfig)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("", "error", err)
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
	user, err := user.Current()
	if err != nil {
		slog.Error("", "error", err)
	}

	return user.HomeDir
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
			slog.Info("Using config file " + config)
			return config, file
		}
		errs = append(errs, err)
	}

	slog.Error("Could not find link config", "errors", errs)
	panic(errs)
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
		slog.Error("Malformed config. Each non-empty line must have exactly two entries.", "line", lineNum)
		panic(1)
	}

	target, err := url.Parse(parts[1])
	if err != nil {
		slog.Error("Malformed config. Invalid url.", "line", lineNum, "url", parts[1])
	}

	return parts[0], target
}

func parseConfig(filePtr *os.File) map[string]url.URL {
	linkMap := make(map[string]url.URL)
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
		linkMap[key] = *target
	}

	return linkMap
}

func (l *LinkMap) watchConfig(watcher *fsnotify.Watcher) {
	dir := filepath.Dir(l.configPath)
	name := filepath.Base(l.configPath)
	err := watcher.Add(dir)
	if err != nil {
		slog.Error("Failed to add watcher on config dir. Config will not live reload", "error", err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			slog.Debug("File watch event received!")
			if filepath.Base(event.Name) == name && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				slog.Debug("Config file updated, reloading...")
				l.update()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			if err != nil {
				slog.Error("File watch error received", "error", err)
			}
			fmt.Println(err)
		}
	}
}

func (l *LinkMap) update() {
	file, err := os.OpenFile(l.configPath, os.O_RDONLY, 0644)
	if err != nil {
		slog.Error("Cound not open config for reading", "file", l.configPath)
		panic(1)
	}
	defer file.Close()

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	slog.Debug("Reading config file", "file", l.configPath)
	l.m = parseConfig(file)
}

func (l *LinkMap) Get(key string) (url.URL, bool) {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()
	target, exists := l.m[key]
	return target, exists
}

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
	l.m[key] = *target

	return nil
}

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

	err = l.replaceConfigInPlace()
	if err != nil {
		return err
	}

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	delete(l.m, key)

	return nil
}

func (l *LinkMap) Update(key string, target *url.URL) error {
	err := l.updateEntry(key, target)
	if err != nil {
		return err
	}

	err = l.replaceConfigInPlace()
	if err != nil {
		return err
	}

	return nil
}

func (l *LinkMap) updateEntry(key string, target *url.URL) error {
	l.fileLock.Lock()
	defer l.fileLock.Unlock()
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

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	l.m[key] = *target

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
