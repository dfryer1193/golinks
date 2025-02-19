package main

import (
	"fmt"
	"github.com/dfryer1193/golinks/config"
	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/mjolnir/router"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.GetConfig()
	logger := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	})
	log.Logger = logger
	zerolog.SetGlobalLevel(cfg.LogLevel)

	r := router.New()
	handler.NewGoLinkService(r, cfg)

	log.Info().Msg("Starting server on port :" + fmt.Sprint(cfg.Port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), r); err != nil {
		log.Fatal().Err(err)
	}
}
