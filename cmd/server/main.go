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

	"agentgo/internal/aggregator"
	"agentgo/internal/config"
	"agentgo/internal/httpserver"
	"agentgo/internal/provider"
	"agentgo/internal/provider/mock"
	simplesummary "agentgo/internal/summary/simple"
)

func main() {
	cfg := config.Load()

	available := map[string]provider.Provider{}
	mockProvider := mock.New()
	available[mockProvider.Name()] = mockProvider

	providers := map[string]provider.Provider{}
	for _, name := range cfg.DefaultProviders {
		if p, ok := available[name]; ok {
			providers[name] = p
		}
	}
	if len(providers) == 0 {
		providers[mockProvider.Name()] = mockProvider
	}

	summ := simplesummary.New()
	agg := aggregator.New(providers, summ, aggregator.Config{
		CacheTTL:       cfg.CacheTTL,
		RequestTimeout: cfg.RequestTimeout,
		HistorySize:    cfg.HistorySize,
	})

	server := httpserver.New(agg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: server.Handler(),
	}

	go func() {
		log.Printf("server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	fmt.Println("server stopped")
}
