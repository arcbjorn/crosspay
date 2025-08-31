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
			"service": "ens-resolver",
			"timestamp": time.Now().Unix(),
		})
	})

	// ENS resolution endpoints
	ensGroup := r.Group("/api/ens")
	{
		ensGroup.GET("/resolve/:name", handleResolveName)
		ensGroup.GET("/reverse/:address", handleReverseResolve)
		ensGroup.POST("/resolve/batch", handleBatchResolve)
		ensGroup.GET("/avatar/:name", handleGetAvatar)
		ensGroup.GET("/text/:name/:key", handleGetTextRecord)
		ensGroup.GET("/search", handleSearchNames)
	}

	// Subname registry endpoints
	subnameGroup := r.Group("/api/subnames")
	{
		subnameGroup.POST("/register", handleRegisterSubname)
		subnameGroup.GET("/list/:domain", handleListSubnames)
		subnameGroup.POST("/bulk", handleBulkRegister)
		subnameGroup.DELETE("/revoke/:subname", handleRevokeSubname)
	}

	// Cache management endpoints
	cacheGroup := r.Group("/api/cache")
	{
		cacheGroup.GET("/stats", handleCacheStats)
		cacheGroup.DELETE("/clear", handleClearCache)
		cacheGroup.DELETE("/entry/:key", handleClearCacheEntry)
	}

	srv := &http.Server{
		Addr:    ":8082",
		Handler: r,
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