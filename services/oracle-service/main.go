package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "oracle-service",
			"timestamp": time.Now().Unix(),
		})
	})

	// FTSO endpoints
	ftsoGroup := r.Group("/api/ftso")
	{
		ftsoGroup.GET("/price/:symbol", handleGetPrice)
		ftsoGroup.GET("/price/:symbol/history", handleGetPriceHistory)
		ftsoGroup.POST("/price/update", handleUpdatePrice)
		ftsoGroup.GET("/symbols", handleGetSupportedSymbols)
	}

	// Random number endpoints
	randomGroup := r.Group("/api/random")
	{
		randomGroup.POST("/request", handleRequestRandom)
		randomGroup.GET("/status/:requestId", handleRandomStatus)
		randomGroup.POST("/fulfill", handleFulfillRandom)
		randomGroup.POST("/winners", handleSelectWinners)
	}

	// FDC endpoints
	fdcGroup := r.Group("/api/fdc")
	{
		fdcGroup.POST("/proof/submit", handleSubmitProof)
		fdcGroup.GET("/proof/verify/:proofId", handleVerifyProof)
		fdcGroup.POST("/proof/confirm", handleConfirmProof)
		fdcGroup.POST("/webhook/payment", handlePaymentWebhook)
		fdcGroup.GET("/proofs", handleGetProofsByTx)
	}

	// Oracle health endpoints
	healthGroup := r.Group("/api/oracle")
	{
		healthGroup.GET("/status", handleOracleStatus)
		healthGroup.POST("/healthcheck", handlePerformHealthCheck)
		healthGroup.POST("/circuit-breaker/pause", handleEmergencyPause)
		healthGroup.POST("/circuit-breaker/resume", handleEmergencyResume)
	}

	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
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

func startHealthMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Starting oracle health monitor...")
	
	for range ticker.C {
		performOracleHealthCheck()
	}
}