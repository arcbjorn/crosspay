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

	"github.com/crosspay/analytics-dashboard/internal/analytics"
	"github.com/crosspay/analytics-dashboard/internal/metrics"
	"github.com/crosspay/analytics-dashboard/internal/websocket"
)

func main() {
	metricsCollector := metrics.NewCollector()
	analyticsService := analytics.NewService(metricsCollector)
	wsHub := websocket.NewHub()

	go wsHub.Run()
	go metricsCollector.StartCollection()

	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /metrics", analyticsService.GetMetrics)
	mux.HandleFunc("GET /metrics/validators", analyticsService.GetValidatorMetrics)
	mux.HandleFunc("GET /metrics/vault", analyticsService.GetVaultMetrics)
	mux.HandleFunc("GET /metrics/payments", analyticsService.GetPaymentMetrics)
	mux.HandleFunc("GET /metrics/privacy", analyticsService.GetPrivacyMetrics)
	mux.HandleFunc("GET /ws", wsHub.HandleWebSocket)
	
	mux.Handle("GET /", http.FileServer(http.Dir("./static/")))

	server := &http.Server{
		Addr:    ":8090",
		Handler: mux,
	}

	go func() {
		log.Println("Starting analytics dashboard on port 8090")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down analytics dashboard...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	wsHub.Stop()
	metricsCollector.Stop()
	log.Println("Analytics dashboard stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "analytics-dashboard",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}