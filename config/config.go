package config

import (
	"flag"
	"fmt"
	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/rs/zerolog"
	"os"
	"strings"
)

type Config struct {
	Port        int
	StorageType storage.StorageType
	ConfigFile  string
	LogLevel    zerolog.Level
}

func help() {
	helptext :=
		`golinks: a simple self-hosted implementation of go links for use in a self-
hosted environment.

Usage: golinks [-port 8080] [-config ./links]

-h                                      Show this help message
-port <number>                          The port to listen on (default: 8080)
-storage <FILE|NONE|SQLITE|POSTGRES>    The type of storage to use for
                                        persistence. Defaults to "FILE". Storage
                                        types:
                                            * NONE: Provides no persistence
                                            * FILE: Persists shortcut entries to
                                                    the file specified by the
                                                    -config option
                                            * SQLITE: Persists shortcut entries
                                                      to a sqlite db. The path
                                                      to the db is specified by
                                                      the -config option.
                                            * POSTGRES: Persists shortcut entries
                                                        to a PostgreSQL database.
                                                        Connection string via
                                                        -config option or
                                                        DATABASE_URL env var.
-config <absolute path to config file>  The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"
                                        For POSTGRES storage, this should be a
                                        connection string (or use DATABASE_URL
                                        environment variable).
-level <loglevel>                       The loglevel to log at. Defaults to
                                        "INFO"

Config format:
The config file is a simple plaintext file consisting of one key/value pair per
line, separated by spaces, like so:

    test https://www.google.com

The value of the pair must be a full web address. Query params are not
respected, though full paths are.`

	fmt.Println(helptext)
	os.Exit(0)
}

func getLevelFromArg(arg string) zerolog.Level {
	switch strings.ToUpper(arg) {
	case "TRACE":
		return zerolog.TraceLevel
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "WARN":
		return zerolog.WarnLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "FATAL":
		return zerolog.FatalLevel
	case "PANIC":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func GetConfig() *Config {
	var port int
	var storageTypeString string
	var configFile string
	var stringLogLevel string
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.StringVar(&storageTypeString, "storage", "FILE", "The type of storage to use for persistence")
	flag.StringVar(&configFile, "config", "", "Location of the config file. Ignored if storageType is 'NONE'")
	flag.StringVar(&stringLogLevel, "level", "INFO", "The level to log at")
	flag.Usage = help

	flag.Parse()

	level, err := zerolog.ParseLevel(stringLogLevel)
	if err != nil {
		fmt.Println("Invalid log level")
		os.Exit(1)
	}

	// For PostgreSQL, use DATABASE_URL env var if config file not specified
	storageType := storage.FromString(storageTypeString)
	if storageType == storage.POSTGRES && configFile == "" {
		configFile = os.Getenv("DATABASE_URL")
		if configFile == "" {
			fmt.Println("PostgreSQL storage requires either -config flag or DATABASE_URL environment variable")
			os.Exit(1)
		}
	}

	return &Config{
		Port:        port,
		StorageType: storageType,
		ConfigFile:  configFile,
		LogLevel:    level,
	}
}
