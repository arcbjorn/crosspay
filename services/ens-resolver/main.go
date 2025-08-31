package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"service": "ens-resolver",
			"timestamp": time.Now().Unix(),
		})
	})

	// ENS resolution endpoints
	mux.HandleFunc("/api/ens/resolve/", handleResolveName)
	mux.HandleFunc("/api/ens/reverse/", handleReverseResolve)
	mux.HandleFunc("/api/ens/resolve/batch", handleBatchResolve)
	mux.HandleFunc("/api/ens/avatar/", handleGetAvatar)
	mux.HandleFunc("/api/ens/text/", handleGetTextRecord)
	mux.HandleFunc("/api/ens/search", handleSearchNames)

	// Subname registry endpoints
	mux.HandleFunc("/api/subnames/register", handleRegisterSubname)
	mux.HandleFunc("/api/subnames/list/", handleListSubnames)
	mux.HandleFunc("/api/subnames/bulk", handleBulkRegister)
	mux.HandleFunc("/api/subnames/revoke/", handleRevokeSubname)

	// Cache management endpoints
	mux.HandleFunc("/api/cache/stats", handleCacheStats)
	mux.HandleFunc("/api/cache/clear", handleClearCache)
	mux.HandleFunc("/api/cache/entry/", handleClearCacheEntry)

	srv := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	// Initialize ENS resolver
	initializeENSResolver()

	go func() {
		log.Println("ENS resolver starting on :8082")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Start background services
	go startCacheEviction()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down ENS resolver...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("ENS resolver stopped")
}

func initializeENSResolver() {
	log.Println("Initializing ENS resolver...")
	
	// Initialize cache
	initCache()
	
	// Initialize ENS client (mock)
	initENSClient()
	
	// Initialize subname registry
	initSubnameRegistry()
	
	log.Println("ENS resolver initialized")
}

func startCacheEviction() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	log.Println("Starting cache eviction process...")
	
	for range ticker.C {
		evictExpiredEntries()
	}
}