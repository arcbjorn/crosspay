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
			"service": "storage-worker",
			"timestamp": time.Now().Unix(),
		})
	})

	// Storage endpoints
	storageGroup := r.Group("/api/storage")
	{
		storageGroup.POST("/upload", handleUpload)
		storageGroup.GET("/retrieve/:cid", handleRetrieve)
		storageGroup.GET("/cost/:size", handleCostEstimate)
	}

	// Receipt endpoints
	receiptGroup := r.Group("/api/receipts")
	{
		receiptGroup.POST("/generate", handleGenerateReceipt)
		receiptGroup.GET("/download/:id", handleDownloadReceipt)
		receiptGroup.GET("/verify/:cid", handleVerifyReceipt)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
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