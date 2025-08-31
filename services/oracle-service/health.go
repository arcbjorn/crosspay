package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type OracleStatus struct {
	Healthy        bool      `json:"healthy"`
	LastCheck      int64     `json:"last_check"`
	Services       ServiceStatus `json:"services"`
	Uptime         int64     `json:"uptime_seconds"`
	Version        string    `json:"version"`
	CircuitBreaker bool      `json:"circuit_breaker_active"`
}

type ServiceStatus struct {
	FTSO   ServiceHealth `json:"ftso"`
	Random ServiceHealth `json:"random"`
	FDC    ServiceHealth `json:"fdc"`
}

type ServiceHealth struct {
	Healthy       bool   `json:"healthy"`
	LastUpdate    int64  `json:"last_update"`
	ErrorCount    int    `json:"error_count"`
	Status        string `json:"status"`
	ResponseTime  int64  `json:"response_time_ms"`
}

var (
	oracleStatus    = &OracleStatus{}
	statusMutex     = sync.RWMutex{}
	startTime       = time.Now()
	circuitBreaker  = false
	healthCheckInterval = 60 * time.Second
)

func initOracleHealth() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	
	oracleStatus = &OracleStatus{
		Healthy:        true,
		LastCheck:      time.Now().Unix(),
		Version:        "1.0.0",
		CircuitBreaker: false,
		Services: ServiceStatus{
			FTSO: ServiceHealth{
				Healthy:      true,
				LastUpdate:   time.Now().Unix(),
				ErrorCount:   0,
				Status:       "operational",
				ResponseTime: 0,
			},
			Random: ServiceHealth{
				Healthy:      true,
				LastUpdate:   time.Now().Unix(),
				ErrorCount:   0,
				Status:       "operational",
				ResponseTime: 0,
			},
			FDC: ServiceHealth{
				Healthy:      true,
				LastUpdate:   time.Now().Unix(),
				ErrorCount:   0,
				Status:       "operational",
				ResponseTime: 0,
			},
		},
	}
}

func performOracleHealthCheck() {
	statusMutex.Lock()
	defer statusMutex.Unlock()
	
	log.Println("Performing oracle health check...")
	
	// Check FTSO health
	ftsoHealth := checkFTSOHealth()
	oracleStatus.Services.FTSO = ftsoHealth
	
	// Check Random service health
	randomHealth := checkRandomHealth()
	oracleStatus.Services.Random = randomHealth
	
	// Check FDC health
	fdcHealth := checkFDCHealth()
	oracleStatus.Services.FDC = fdcHealth
	
	// Overall health is true if all services are healthy and circuit breaker is off
	overallHealth := ftsoHealth.Healthy && randomHealth.Healthy && fdcHealth.Healthy && !circuitBreaker
	
	oracleStatus.Healthy = overallHealth
	oracleStatus.LastCheck = time.Now().Unix()
	oracleStatus.Uptime = int64(time.Since(startTime).Seconds())
	oracleStatus.CircuitBreaker = circuitBreaker
	
	if overallHealth {
		log.Println("Oracle health check passed - all services operational")
	} else {
		log.Printf("Oracle health check failed - FTSO: %t, Random: %t, FDC: %t, Circuit Breaker: %t", 
			ftsoHealth.Healthy, randomHealth.Healthy, fdcHealth.Healthy, circuitBreaker)
	}
}

func checkFTSOHealth() ServiceHealth {
	start := time.Now()
	
	// Check if we have recent price updates
	pricesMutex.RLock()
	recentPrices := 0
	now := time.Now().Unix()
	
	for _, priceData := range currentPrices {
		if now - priceData.Timestamp < 300 { // 5 minutes
			recentPrices++
		}
	}
	pricesMutex.RUnlock()
	
	responseTime := time.Since(start).Milliseconds()
	healthy := recentPrices >= len(supportedSymbols)/2 // At least half the symbols should have recent data
	
	var status string
	var errorCount int
	
	if healthy {
		status = "operational"
		errorCount = 0
	} else {
		status = "degraded"
		errorCount = len(supportedSymbols) - recentPrices
	}
	
	return ServiceHealth{
		Healthy:      healthy,
		LastUpdate:   now,
		ErrorCount:   errorCount,
		Status:       status,
		ResponseTime: responseTime,
	}
}

func checkRandomHealth() ServiceHealth {
	start := time.Now()
	
	// Check pending random requests
	randomMutex.RLock()
	pendingCount := 0
	overdueCount := 0
	now := time.Now().Unix()
	
	for _, request := range randomRequests {
		if request.Status == "pending" {
			pendingCount++
			if now - request.Timestamp > 300 { // 5 minutes overdue
				overdueCount++
			}
		}
	}
	randomMutex.RUnlock()
	
	responseTime := time.Since(start).Milliseconds()
	healthy := overdueCount == 0 && pendingCount < 100 // No overdue requests and not too many pending
	
	var status string
	
	if healthy {
		status = "operational"
	} else if overdueCount > 0 {
		status = "degraded"
	} else {
		status = "overloaded"
	}
	
	return ServiceHealth{
		Healthy:      healthy,
		LastUpdate:   now,
		ErrorCount:   overdueCount,
		Status:       status,
		ResponseTime: responseTime,
	}
}

func checkFDCHealth() ServiceHealth {
	start := time.Now()
	
	// Check recent proof submissions
	proofsMutex.RLock()
	recentProofs := 0
	failedProofs := 0
	now := time.Now().Unix()
	
	for _, proof := range externalProofs {
		if now - proof.Timestamp < 3600 { // 1 hour
			recentProofs++
			if proof.Status == "rejected" {
				failedProofs++
			}
		}
	}
	proofsMutex.RUnlock()
	
	responseTime := time.Since(start).Milliseconds()
	
	// FDC is healthy if we don't have too many failed proofs
	failureRate := float64(failedProofs) / float64(recentProofs + 1) // +1 to avoid division by zero
	healthy := failureRate < 0.5 // Less than 50% failure rate
	
	var status string
	
	if healthy {
		status = "operational"
	} else {
		status = "degraded"
	}
	
	return ServiceHealth{
		Healthy:      healthy,
		LastUpdate:   now,
		ErrorCount:   failedProofs,
		Status:       status,
		ResponseTime: responseTime,
	}
}

func handleOracleStatus(c *gin.Context) {
	statusMutex.RLock()
	status := *oracleStatus // Copy the status
	statusMutex.RUnlock()
	
	c.JSON(http.StatusOK, status)
}

func handlePerformHealthCheck(c *gin.Context) {
	go performOracleHealthCheck()
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Health check initiated",
		"timestamp": time.Now().Unix(),
	})
}

func handleEmergencyPause(c *gin.Context) {
	statusMutex.Lock()
	circuitBreaker = true
	oracleStatus.Healthy = false
	oracleStatus.CircuitBreaker = true
	statusMutex.Unlock()
	
	log.Println("EMERGENCY: Oracle services paused via circuit breaker")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Oracle services paused",
		"circuit_breaker": true,
		"timestamp": time.Now().Unix(),
	})
}

func handleEmergencyResume(c *gin.Context) {
	statusMutex.Lock()
	circuitBreaker = false
	oracleStatus.CircuitBreaker = false
	statusMutex.Unlock()
	
	// Perform immediate health check after resuming
	go performOracleHealthCheck()
	
	log.Println("Oracle services resumed, performing health check...")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Oracle services resumed",
		"circuit_breaker": false,
		"timestamp": time.Now().Unix(),
	})
}

// Middleware to check if oracle is healthy before processing requests
func requireHealthyOracle() gin.HandlerFunc {
	return func(c *gin.Context) {
		statusMutex.RLock()
		healthy := oracleStatus.Healthy && !circuitBreaker
		statusMutex.RUnlock()
		
		if !healthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Oracle services unavailable",
				"circuit_breaker_active": circuitBreaker,
				"retry_after_seconds": 60,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func startHealthMonitor() {
	// Initialize health status
	initOracleHealth()
	
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()
	
	log.Printf("Starting oracle health monitor (interval: %v)", healthCheckInterval)
	
	for {
		select {
		case <-ticker.C:
			performOracleHealthCheck()
		}
	}
}