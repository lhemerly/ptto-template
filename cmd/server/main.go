package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lhemerly/ptto-template-go/internal/app"
)

func main() {
	if err := run(); err != nil {
		log.Printf("fatal error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New()
	if err != nil {
		return fmt.Errorf("startup failed: %w", err)
	}
	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			log.Printf("shutdown close error: %v", closeErr)
		}
	}()

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      application.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			log.Printf("http shutdown error: %v", shutdownErr)
		}
	}()

	log.Printf("listening on http://localhost%s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}

	return nil
}
