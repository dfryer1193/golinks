package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dfryer1193/golinks/config"
	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/dfryer1193/mjolnir/router"
	"net/http"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.GetConfig()
	zerolog.SetGlobalLevel(cfg.LogLevel)

	// Handle migration if requested
	if cfg.MigrateFrom != "" {
		if err := runMigration(cfg); err != nil {
			log.Fatal().Err(err).Msg("Migration failed")
		}
		return
	}

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

func runMigration(cfg *config.Config) error {
	if cfg.StorageType != storage.SQLITE && cfg.StorageType != storage.POSTGRES {
		return fmt.Errorf("--migrate-from requires storage type SQLITE or POSTGRES, got %s", cfg.StorageType)
	}

	var target storage.Storage
	var err error

	switch cfg.StorageType {
	case storage.SQLITE:
		target, err = storage.NewSQLiteStorage(cfg.ConfigFile)
	case storage.POSTGRES:
		target, err = storage.NewPostgresStorage(cfg.ConfigFile)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize target storage: %w", err)
	}

	return storage.MigrateFromFile(cfg.MigrateFrom, target)
}