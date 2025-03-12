package storage

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

type FileStorage struct {
	configPath    string
	fileLock      *sync.RWMutex
	watcher       *fsnotify.Watcher
	reloadChannel chan bool
}

func NewFileStorage(configPath string) *FileStorage {
	path := findConfig(configPath)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating file watcher")
	}

	storage := &FileStorage{
		configPath:    path,
		fileLock:      &sync.RWMutex{},
		watcher:       watcher,
		reloadChannel: make(chan bool),
	}

	go storage.watchConfig()

	return storage
}

func (f *FileStorage) watchConfig() {
	dir := filepath.Dir(f.configPath)
	name := filepath.Base(f.configPath)
	err := f.watcher.Add(dir)
	if err != nil {
		log.Err(err).Msg("Failed to add watcher on config dir. Config will not live reload")
	}
	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) == name && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				f.reloadChannel <- true
			}
		case err, ok := <-f.watcher.Errors:
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

func getHomeDir() string {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get current user")
	}

	return currentUser.HomeDir
}

func findConfig(requestedConfig string) string {
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
			log.Info().Str("file", config).Msg("Found config")
			file.Close()
			return config
		}
		errs = append(errs, err)
	}

	for _, err := range errs {
		log.Err(err)
	}
	log.Fatal().Msg("Failed to find any config")
	return ""
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

func (f *FileStorage) Read() map[string]string {
	linkMap := make(map[string]string)
	filePtr, err := openFile(f.configPath)
	if err != nil {
		log.Error().Err(err).Str("file path", f.configPath).Msg("Failed to open file for reading")
		return linkMap
	}
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

// Put appends a new entry to the link config. If the entry already exists, it will be duplicated in the file.
func (f *FileStorage) Put(key string, target string) {
	f.fileLock.Lock()
	defer f.fileLock.Unlock()

	file, err := os.OpenFile(f.configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("file path", f.configPath).
			Msg("Failed to open file for writing")
		return
	}
	defer file.Close()

	if _, err := file.WriteString(key + " " + target + "\n"); err != nil {
		log.
			Error().
			Err(err).
			Str("file path", f.configPath).
			Str("key", key).
			Str("target", target).
			Msg("Failed to write to file")
	}

}

func (f *FileStorage) Delete(key string) {
	changed, err := f.updateEntry(key, "")
	if err != nil {
		log.
			Error().
			Err(err).
			Str("key", key).
			Msg("Failed to delete key")
	}

	if changed {
		err = f.replaceConfigInPlace()
		if err != nil {
			log.
				Error().
				Err(err).
				Str("key", key).
				Msg("Failed to replace config file in place after delete")
		}
	}
}

func (f *FileStorage) Update(key string, target string) {
	changed, err := f.updateEntry(key, target)
	if err != nil {
		log.
			Error().
			Err(err).
			Str("key", key).
			Str("target", target).
			Msg("Failed to update key")
	}

	if changed {
		err = f.replaceConfigInPlace()
		if err != nil {
			log.
				Error().
				Err(err).
				Str("key", key).
				Str("target", target).
				Msg("Failed to replace config file in place after update")
		}
	}
}

func (f *FileStorage) updateEntry(key string, target string) (bool, error) {
	f.fileLock.Lock()
	defer f.fileLock.Unlock()

	curFile, err := os.OpenFile(f.configPath, os.O_RDONLY, 0600)
	if err != nil {
		return false, err
	}
	defer curFile.Close()

	newFile, err := os.OpenFile(f.getScratchConfigFilepath(), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false, err
	}
	defer newFile.Close()

	var changed = false
	scanner := bufio.NewScanner(curFile)
	for scanner.Scan() {
		txt := scanner.Text()

		// Path exists somewhere in the file
		if strings.HasPrefix(txt, key+" ") {
			if target == "" {
				changed = true
				continue
			}

			if _, err := newFile.WriteString(key + " " + target + "\n"); err != nil {
				return false, err
			}
			changed = true
			continue
		}

		// If the path doesn't match, just write the line out
		_, err := newFile.WriteString(txt + "\n")
		if err != nil {
			fmt.Println(err)
			return false, err
		}
	}

	return changed, nil
}

func (f *FileStorage) getScratchConfigFilepath() string {
	return f.configPath + "~"
}

func (f *FileStorage) getBackupConfigFilepath() string {
	return f.configPath + ".bak"
}

func (f *FileStorage) replaceConfigInPlace() error {
	f.fileLock.Lock()
	defer f.fileLock.Unlock()
	err := os.Rename(f.configPath, f.getBackupConfigFilepath())
	if err != nil {
		return err
	}
	err = os.Rename(f.getScratchConfigFilepath(), f.configPath)
	if err != nil {
		return err
	}
	err = os.Remove(f.getBackupConfigFilepath())
	if err != nil {
		return err
	}
	return nil
}

func (f *FileStorage) GetReloadChannel() <-chan bool {
	return f.reloadChannel
}
