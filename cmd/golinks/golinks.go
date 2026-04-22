package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dfryer1193/golinks/config"
	"github.com/dfryer1193/golinks/internal/handler"
	"github.com/dfryer1193/golinks/internal/links/storage"
	"github.com/dfryer1193/golinks/internal/migrations"
	"github.com/dfryer1193/mjolnir/router"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		// Check if second argument is 'status' subcommand (positional)
		// Usage: golinks migrate [status] -storage <type> -config <path>
		statusCmd := false
		if len(os.Args) > 2 && os.Args[2] == "status" {
			statusCmd = true
			// Remove both 'migrate' and 'status' from os.Args
			// This leaves: [golinks, -storage, SQLITE, -config, test.db]
			os.Args = append([]string{os.Args[0]}, os.Args[3:]...)
		} else {
			// Remove just 'migrate' from os.Args
			// This leaves: [golinks, -storage, SQLITE, -config, test.db]
			os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		}
		runMigration(statusCmd)
		return
	}

	cfg := config.GetConfig()
	zerolog.SetGlobalLevel(cfg.LogLevel)

	r := router.New()
	service := handler.NewGoLinkService(r, cfg)
	defer func() {
		if err := service.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close service")
		}
	}()

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

func runMigration(statusCmd bool) {
	cfg := config.GetConfig()
	
	// Auto-detect database type and connection string
	var dbType, connStr string
	
	if cfg.StorageType == storage.SQLITE {
		dbType = "sqlite"
		connStr = cfg.ConfigFile
		if connStr == "" {
			fmt.Println("Error: SQLite requires -config flag or config file path")
			os.Exit(1)
		}
	} else if cfg.StorageType == storage.POSTGRES {
		dbType = "postgres"
		connStr = cfg.ConfigFile
		if connStr == "" {
			connStr = os.Getenv("DATABASE_URL")
		}
		if connStr == "" {
			fmt.Println("Error: PostgreSQL requires -config flag or DATABASE_URL environment variable")
			os.Exit(1)
		}
	} else {
		fmt.Println("Error: Migration only supports SQLITE and POSTGRES storage types")
		fmt.Println("Usage: golinks migrate -storage <SQLITE|POSTGRES> [-config <path_or_connection_string>]")
		os.Exit(1)
	}

	// Open database connection
	var db *sql.DB
	var err error
	
	if dbType == "sqlite" {
		db, err = sql.Open("sqlite3", connStr)
	} else {
		db, err = sql.Open("postgres", connStr)
	}
	
	if err != nil {
		fmt.Printf("Error: Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Verify connection
	if err := db.Ping(); err != nil {
		fmt.Printf("Error: Failed to ping database: %v\n", err)
		os.Exit(1)
	}

	// Create migrator
	migrator, err := migrations.NewMigrator(db, dbType)
	if err != nil {
		fmt.Printf("Error: Failed to create migrator: %v\n", err)
		os.Exit(1)
	}

	// Check for status flag
	if statusCmd {
		if err := migrator.Status(); err != nil {
			fmt.Printf("Error: Failed to get migration status: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Run migrations
	fmt.Printf("Running migrations for %s...\n", dbType)
	if err := migrator.Migrate(); err != nil {
		fmt.Printf("Error: Migration failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✓ Migrations completed successfully")
}
