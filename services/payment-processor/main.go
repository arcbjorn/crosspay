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
	mux.HandleFunc("/health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "payment-processor",
			"timestamp": time.Now().Unix(),
		})
	}))

	// Payment API endpoints
	mux.HandleFunc("/api/payments/create", corsHandler(handleCreatePayment))
	mux.HandleFunc("/api/payments/complete/", corsHandler(handleCompletePayment))
	mux.HandleFunc("/api/payments/refund/", corsHandler(handleRefundPayment))
	mux.HandleFunc("/api/payments/", corsHandler(handleGetPayment))
	mux.HandleFunc("/api/payments/user/", corsHandler(handleGetUserPayments))

	// Receipt API endpoints
	mux.HandleFunc("/api/receipts/generate/", corsHandler(handleGenerateReceipt))
	mux.HandleFunc("/api/receipts/download/", corsHandler(handleDownloadReceipt))
	mux.HandleFunc("/api/receipts/verify/", corsHandler(handleVerifyReceipt))
	mux.HandleFunc("/api/receipts/payment/", corsHandler(handleGetReceiptsByPayment))

	// Oracle integration endpoints
	mux.HandleFunc("/api/oracle/price/", corsHandler(handleGetPrice))
	mux.HandleFunc("/api/oracle/random/request", corsHandler(handleRequestRandom))
	mux.HandleFunc("/api/oracle/random/status/", corsHandler(handleRandomStatus))
	mux.HandleFunc("/api/oracle/proof/submit", corsHandler(handleSubmitProof))
	mux.HandleFunc("/api/oracle/proof/verify/", corsHandler(handleVerifyProof))

	// ENS resolution endpoints
	mux.HandleFunc("/api/ens/resolve/", corsHandler(handleResolveName))
	mux.HandleFunc("/api/ens/reverse/", corsHandler(handleReverseResolve))
	mux.HandleFunc("/api/ens/resolve/batch", corsHandler(handleBatchResolve))

	// Storage endpoints
	mux.HandleFunc("/api/storage/upload", corsHandler(handleUploadFile))
	mux.HandleFunc("/api/storage/retrieve/", corsHandler(handleRetrieveFile))
	mux.HandleFunc("/api/storage/cost/", corsHandler(handleEstimateCost))

	// Analytics endpoints
	mux.HandleFunc("/api/analytics/stats", corsHandler(handleGetStats))
	mux.HandleFunc("/api/analytics/payments/volume", corsHandler(handleGetPaymentVolume))
	mux.HandleFunc("/api/analytics/receipts/stats", corsHandler(handleGetReceiptStats))

	srv := &http.Server{
		Addr:    ":8083",
		Handler: mux,
	}

	// Initialize services
	initializeServices()

	go func() {
		log.Println("Payment processor starting on :8083")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down payment processor...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Payment processor stopped")
}

func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next(w, r)
	}
}

func initializeServices() {
	log.Println("Initializing payment processor services...")
	
	// Initialize service clients
	initStorageClient()
	initOracleClient() 
	initENSClient()
	
	// Initialize database
	initDatabase()
	
	log.Println("Payment processor services initialized")
}