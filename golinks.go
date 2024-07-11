package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/golinks/internal/links"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func help() {
	helptext :=
		`golinks: a simple self-hosted implementation of go links for use in a self-
hosted environment.

Usage: golinks [-port 8080] [-config ./links]

-h                                      Show this help message
-port <number>                          The port to listen on (default: 8080)
-config <absolute path to config file>  The path to the preferred config file.
                                        If this file is not present, falls back
                                        to default locations in the following
                                        order:
                                            * "./links"
                                            * "~/.config/golinks/links"
                                            * "/etc/golinks/links"

Config format:
The config file is a simple plaintext file consisting of one key/value pair per
line, separated by spaces, like so:
	
    test https://www.google.com

The value of the pair must be a full web address. Query params are not
respected, though full paths are.
`

	fmt.Println(helptext)
	os.Exit(0)
}

func main() {
	var port int
	var configFile string
	var stringLogLevel string
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.StringVar(&configFile, "config", "", "Location of the config file")
	flag.StringVar(&stringLogLevel, "level", "INFO", "The level to log at")
	flag.Usage = help

	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	})
	level := getLevelFromArg(stringLogLevel)
	log.Info().Str("loglevel", level.String()).Msg("Setting log level...")
	zerolog.SetGlobalLevel(level)

	log.Info().Int("port", port).Msg("Starting http server")

	redirector := handler.NewGolinkHandler(links.NewLinkMap(configFile))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), redirector); err != nil {
		log.Fatal().Err(err)
	}
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
