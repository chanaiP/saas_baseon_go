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

	"saas_baseon_go/internal/bootstrap"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("api stopped: %v", err)
	}
}

func run() error {
	cfg := bootstrap.LoadConfig()

	db, err := bootstrap.NewPostgres(cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	redisClient := bootstrap.NewRedis(cfg)
	router := bootstrap.NewRouter(cfg, db, redisClient)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("saas baseon go api listening on %s", cfg.HTTPAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-stop:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(ctx)
	}
}
