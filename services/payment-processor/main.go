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
	
	// CORS middleware
	r.Use(corsMiddleware())
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "payment-processor",
			"timestamp": time.Now().Unix(),
		})
	})

	// Payment API endpoints
	paymentGroup := r.Group("/api/payments")
	{
		paymentGroup.POST("/create", handleCreatePayment)
		paymentGroup.POST("/complete/:id", handleCompletePayment)
		paymentGroup.POST("/refund/:id", handleRefundPayment)
		paymentGroup.GET("/:id", handleGetPayment)
		paymentGroup.GET("/user/:address", handleGetUserPayments)
	}

	// Receipt API endpoints
	receiptGroup := r.Group("/api/receipts")
	{
		receiptGroup.POST("/generate/:paymentId", handleGenerateReceipt)
		receiptGroup.GET("/download/:id", handleDownloadReceipt)
		receiptGroup.GET("/verify/:cid", handleVerifyReceipt)
		receiptGroup.GET("/payment/:paymentId", handleGetReceiptsByPayment)
	}

	// Oracle integration endpoints
	oracleGroup := r.Group("/api/oracle")
	{
		oracleGroup.GET("/price/:symbol", handleGetPrice)
		oracleGroup.POST("/random/request", handleRequestRandom)
		oracleGroup.GET("/random/status/:requestId", handleRandomStatus)
		oracleGroup.POST("/proof/submit", handleSubmitProof)
		oracleGroup.GET("/proof/verify/:proofId", handleVerifyProof)
	}

	// ENS resolution endpoints
	ensGroup := r.Group("/api/ens")
	{
		ensGroup.GET("/resolve/:name", handleResolveName)
		ensGroup.GET("/reverse/:address", handleReverseResolve)
		ensGroup.POST("/resolve/batch", handleBatchResolve)
	}

	// Storage endpoints
	storageGroup := r.Group("/api/storage")
	{
		storageGroup.POST("/upload", handleUploadFile)
		storageGroup.GET("/retrieve/:cid", handleRetrieveFile)
		storageGroup.GET("/cost/:size", handleEstimateCost)
	}

	// Analytics endpoints
	analyticsGroup := r.Group("/api/analytics")
	{
		analyticsGroup.GET("/stats", handleGetStats)
		analyticsGroup.GET("/payments/volume", handleGetPaymentVolume)
		analyticsGroup.GET("/receipts/stats", handleGetReceiptStats)
	}

	srv := &http.Server{
		Addr:    ":8083",
		Handler: r,
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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
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