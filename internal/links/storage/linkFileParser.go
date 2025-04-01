package storage

import (
	"bufio"
	"io"
	"net/url"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"
)

func parseLinksFile(reader io.Reader) (map[string]string, error) {
	newLinks := make(map[string]string)
	sc := bufio.NewScanner(reader)
	lineNum := 0

	for sc.Scan() {
		lineNum++
		line := sc.Text()
		key, target, err := parseLine(line, lineNum)
		if err != nil {
			return nil, err
		}
		if key == "" && target == nil {
			continue
		}
		newLinks[key] = target.String()
	}

	return newLinks, nil
}

func parseLine(line string, lineNum int) (string, *url.URL, error) {
	parts := strings.FieldsFunc(strings.TrimSpace(line), func(c rune) bool { return unicode.IsSpace(c) })
	if len(parts) != 2 {
		if len(parts) == 0 {
			return "", nil, nil
		}
		err := &ParseError{}
		log.Error().Err(err).Int("line", lineNum).Msg("Malformed config. Each non-empty line must have exactly two entries.")
		return "", nil, err
	}

	target, err := url.Parse(parts[1])
	if err != nil {
		log.Err(err).Int("line", lineNum).Str("url", parts[1]).Msg("Malformed config. Invalid url")
		return "", nil, &ParseError{}
	}

	return parts[0], target, nil
}
