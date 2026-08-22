package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"botmaker-backend/config"
	"botmaker-backend/internal/engine"
	"botmaker-backend/internal/handlers"
	"botmaker-backend/internal/storage"
	"botmaker-backend/internal/templates"
)

func main() {
	cfg := config.Load()

	store, err := storage.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	defer store.Close()

	eng := engine.New(store, cfg)
	eng.RegisterTemplate(&templates.AniTez{})
	eng.RegisterTemplate(&templates.AniXUltra{})

	if err := eng.Start(); err != nil {
		log.Fatalf("engine: %v", err)
	}

	api := handlers.NewAPI(eng, store)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("botmaker platform listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("shutting down…")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
