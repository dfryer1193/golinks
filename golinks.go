package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
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
	flag.IntVar(&port, "port", 8080, "The port to listen on")
	flag.StringVar(&configFile, "config", "", "Location of the config file")
	flag.Usage = help

	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	})

	log.Info().Int("port", port).Msg("Starting http server")

	redirector := handler.NewGolinkHandler(links.NewLinkMap(configFile))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), redirector); err != nil {
		log.Fatal().Err(err)
	}
}
