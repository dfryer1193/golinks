package main

import (
	"fmt"
	"github.com/dfryer1193/golinks/config"
	"net/http"
	"os"
	"time"

	"github.com/dfryer1193/golinks/internal/links/storage"

	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/golinks/internal/links"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho/v3"
)

func main() {
	cfg := config.GetConfig()
	logger := log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	})
	log.Logger = logger
	zerolog.SetGlobalLevel(cfg.LogLevel)

	service := echo.New()
	service.Logger = lecho.From(logger)
	service.Use(middleware.Logger())
	if err := service.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal().Err(err)
	}

	redirector := handler.NewGoLinkService(links.NewLinkMap(storageType, configFile))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), redirector); err != nil {
		log.Fatal().Err(err)
	}
}
