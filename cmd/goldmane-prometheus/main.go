package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danielgo/goldmane-prometheus/internal/collector"
	"github.com/danielgo/goldmane-prometheus/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	log.Println("Starting Goldmane Prometheus Exporter")

	// Load configuration
	cfg := config.LoadFromEnv()
	log.Printf("Configuration loaded: GoldmaneAddr=%s, MetricsAddr=%s, PollInterval=%s",
		cfg.GoldmaneAddr, cfg.MetricsAddr, cfg.PollInterval)

	// Create collector
	col, err := collector.NewCollector(cfg)
	if err != nil {
		log.Fatalf("Failed to create collector: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Goldmane API
	if err := col.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to Goldmane API: %v", err)
	}
	defer col.Close()

	// Start metrics HTTP server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	server := &http.Server{
		Addr:         cfg.MetricsAddr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("Starting metrics server on %s", cfg.MetricsAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Start flow collector in goroutine
	go func() {
		if err := col.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Flow collector stopped with error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutdown signal received, gracefully shutting down...")

	// Cancel context to stop collector
	cancel()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("Shutdown complete")
}
