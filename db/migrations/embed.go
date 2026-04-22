package migrations

import "embed"

//go:embed sqlite/*.sql
var SqliteMigrations embed.FS

//go:embed postgres/*.sql
var PostgresMigrations embed.FS
