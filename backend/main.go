package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"printsec-warroom/backend/internal/api"
	"printsec-warroom/backend/internal/config"
	"printsec-warroom/backend/internal/db"
	"printsec-warroom/backend/internal/repository"
)

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close()

	if cfg.AutoMigrate {
		path := filepath.Join("migrations", "001_init.sql")
		if err := db.Migrate(ctx, conn, path); err != nil {
			log.Fatalf("migrate database: %v", err)
		}
	}

	repo := repository.New(conn)
	app := api.New(repo, cfg)
	server := &http.Server{
		Addr:              cfg.Address(),
		Handler:           app.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("PrintSec War Room API listening on http://%s", cfg.Address())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
