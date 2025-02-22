package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dfryer1193/golinks/config"
	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/mjolnir/router"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.GetConfig()
	zerolog.SetGlobalLevel(cfg.LogLevel)

	r := router.New()
	handler.NewGoLinkService(r, cfg)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: r,
	}

	go func() {
		log.Info().Msg("Starting server on port :" + fmt.Sprint(cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown server")
	}

	log.Info().Msg("Server stopped")
}
