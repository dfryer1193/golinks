package links

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
)

func NewLinkMap(requestedConfig string) map[string]url.URL {
	config := findConfig(requestedConfig)
	return parseConfig(config)
}

func findConfig(requestedConfig string) *os.File {
	configs := []string{
		"./links",
		"~/.config/golinks/links",
		"/etc/golinks/config",
	}
	errs := []error{}
	if requestedConfig != "" {
		file, err := os.OpenFile(requestedConfig, os.O_RDONLY, 0644)
		if err == nil {
			fmt.Printf("Using config file %s\n", requestedConfig)
			return file
		}
		errs = append(errs, err)
	}

	for _, config := range configs {
		file, err := os.OpenFile(config, os.O_RDONLY, 0644)
		if err == nil {
			fmt.Printf("Using config file %s", config)
			return file
		}
		errs = append(errs, err)
	}

	log.Fatal("Could not find link config. Errors: ", errs)
	return nil
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
