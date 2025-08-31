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
	log.Println("Starting CrossPay Storage Worker...")

	// Initialize SynapseSDK client
	initStorage()

	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"service": "storage-worker",
			"timestamp": time.Now().Unix(),
		})
	}))

	// Storage endpoints
	mux.HandleFunc("/api/storage/upload", corsHandler(handleUpload))
	mux.HandleFunc("/api/storage/retrieve/", corsHandler(handleRetrieve))
	mux.HandleFunc("/api/storage/cost/", corsHandler(handleCostEstimate))
	mux.HandleFunc("/api/storage/files", corsHandler(handleListFiles))
	mux.HandleFunc("/api/storage/pin/", corsHandler(handlePinToIPFS))
	mux.HandleFunc("/api/storage/deal-status/", corsHandler(handleDealStatus))
	mux.HandleFunc("/api/storage/network/info", corsHandler(handleNetworkInfo))

	// Receipt endpoints
	mux.HandleFunc("/api/receipts/generate", corsHandler(handleGenerateReceipt))
	mux.HandleFunc("/api/receipts/download/", corsHandler(handleDownloadReceipt))
	mux.HandleFunc("/api/receipts/verify/", corsHandler(handleVerifyReceipt))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Println("Storage worker starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down storage worker...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Storage worker stopped")
}

func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next(w, r)
	}
}