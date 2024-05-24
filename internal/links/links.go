package links

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

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
		log.Fatal(err)
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
		log.Fatal(err)
	}

	return user.HomeDir
}

func findConfig(requestedConfig string) (string, *os.File) {
	homedir := getHomeDir()
	configs := []string{
		"./links",
		homedir + "/.config/golinks/links",
		"/etc/golinks/links",
	}
	errs := []error{}
	if requestedConfig != "" {
		fpath, err := filepath.Abs(requestedConfig)
		if err == nil {
			file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
			if err == nil {
				fmt.Printf("Using config file %s\n", requestedConfig)
				return requestedConfig, file
			}
		}
		errs = append(errs, err)
	}

	for _, config := range configs {
		fpath, err := filepath.Abs(requestedConfig)
		if err == nil {

			file, err := os.OpenFile(fpath, os.O_RDONLY, 0644)
			if err == nil {
				fmt.Printf("Using config file %s\n", config)
				return config, file
			}
		}
		errs = append(errs, err)
	}

	log.Fatal("Could not find link config. Errors: ", errs)
	return "", nil // unreachable code
}

func parseLine(line string, lineNum int) (string, *url.URL) {
	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) != 2 {
		log.Fatalf("Malformed config (L%d): each line must have exactly two entries.", lineNum)
	}

	target, err := url.Parse(parts[1])
	if err != nil {
		log.Fatalf("Malformed config (L%d): invalid URL \"%s\"", lineNum, parts[1])
	}

	return parts[0], target
}

func parseConfig(filePtr *os.File) map[string]url.URL {
	// TODO: Ensure file ends with a newline
	linkMap := make(map[string]url.URL)
	defer filePtr.Close()

	scanner := bufio.NewScanner(filePtr)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		key, target := parseLine(scanner.Text(), lineNum)
		linkMap[key] = *target
	}

	return linkMap
}

func (l *LinkMap) watchConfig(watcher *fsnotify.Watcher) {
	dir := filepath.Dir(l.configPath)
	name := filepath.Base(l.configPath)
	err := watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			if event.Name == name && (event.Has(fsnotify.Write) || event.Has(fsnotify.Create)) {
				fmt.Println("Updating...")
				l.update()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println(err)
		}
	}
}

func (l *LinkMap) update() {
	file, err := os.OpenFile(l.configPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatalf("Could not open config %s for reading.", l.configPath)
	}
	defer file.Close()

	l.mapLock.Lock()
	defer l.mapLock.Unlock()
	l.m = parseConfig(file)
}

func (l *LinkMap) Get(key string) (url.URL, bool) {
	l.mapLock.RLock()
	defer l.mapLock.RUnlock()
	target, exists := l.m[key]
	return target, exists
}

func (l *LinkMap) Put(key string, target url.URL) error {
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

	return nil
}
