package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

func MigrateFromFile(sourcePath string, target Storage) error {
	if target == nil {
		return fmt.Errorf("target storage is nil")
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer file.Close()

	newLinks, err := parseLinksFile(file)
	if err != nil {
		return fmt.Errorf("failed to parse source file: %w", err)
	}

	if len(newLinks) == 0 {
		log.Info().Msg("No links found in source file to migrate")
		return nil
	}

	existingLinks, err := target.Read()
	if err != nil {
		return fmt.Errorf("failed to read existing links: %w", err)
	}

	conflicts := 0
	for k, v := range newLinks {
		if existing, ok := existingLinks[k]; ok && existing != v {
			conflicts++
		}
	}
	if conflicts > 0 {
		log.Info().Int("conflicts", conflicts).Msg("Key conflicts detected; new values will overwrite")
	}

	var sb strings.Builder
	for k, v := range newLinks {
		sb.WriteString(k)
		sb.WriteString(" ")
		sb.WriteString(v)
		sb.WriteString("\n")
	}

	_, err = target.ReplaceConfig(strings.NewReader(sb.String()))
	if err != nil {
		return fmt.Errorf("failed to write migrated links: %w", err)
	}

	log.Info().Int("count", len(newLinks)).Msg("Migration completed successfully")
	return nil
}