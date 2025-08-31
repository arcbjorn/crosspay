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
			"service": "oracle-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// FTSO endpoints
	mux.HandleFunc("/api/ftso/price/", handleGetPrice)
	mux.HandleFunc("/api/ftso/symbols", handleGetSupportedSymbols)
	mux.HandleFunc("/api/ftso/price/update", handleUpdatePrice)

	// Random number endpoints
	mux.HandleFunc("/api/random/request", handleRequestRandom)
	mux.HandleFunc("/api/random/status/", handleRandomStatus)
	mux.HandleFunc("/api/random/fulfill", handleFulfillRandom)
	mux.HandleFunc("/api/random/winners", handleSelectWinners)

	// FDC endpoints
	mux.HandleFunc("/api/fdc/proof/submit", handleSubmitProof)
	mux.HandleFunc("/api/fdc/proof/verify/", handleVerifyProof)
	mux.HandleFunc("/api/fdc/proof/confirm", handleConfirmProof)
	mux.HandleFunc("/api/fdc/webhook/payment", handlePaymentWebhook)
	mux.HandleFunc("/api/fdc/proofs", handleGetProofsByTx)

	// Oracle health endpoints
	mux.HandleFunc("/api/oracle/status", handleOracleStatus)
	mux.HandleFunc("/api/oracle/healthcheck", handlePerformHealthCheck)
	mux.HandleFunc("/api/oracle/circuit-breaker/pause", handleEmergencyPause)
	mux.HandleFunc("/api/oracle/circuit-breaker/resume", handleEmergencyResume)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	// Initialize oracle services
	initializeOracle()

	go func() {
		log.Println("Oracle service starting on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Start background services
	go startPriceFeedUpdater()
	go startRandomFulfiller()
	go startHealthMonitor()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down oracle service...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Oracle service stopped")
}

func initializeOracle() {
	log.Println("Initializing oracle services...")
	
	// Initialize FTSO client (mock)
	initializeFTSO()
	
	// Initialize RNG client (mock)
	initializeRNG()
	
	// Initialize FDC client (mock)
	initializeFDC()
	
	log.Println("Oracle services initialized")
}

func startPriceFeedUpdater() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("Starting price feed updater...")
	
	for range ticker.C {
		updatePriceFeeds()
	}
}

func startRandomFulfiller() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("Starting random number fulfiller...")
	
	for range ticker.C {
		fulfillPendingRandomRequests()
	}
}

