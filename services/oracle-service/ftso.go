package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type PriceData struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	Decimals  int     `json:"decimals"`
	Valid     bool    `json:"valid"`
}

type PriceHistory struct {
	Symbol string      `json:"symbol"`
	Data   []PriceData `json:"data"`
}

var (
	currentPrices = make(map[string]PriceData)
	priceHistory  = make(map[string][]PriceData)
	pricesMutex   = sync.RWMutex{}
	
	supportedSymbols = []string{
		"ETH/USD", "BTC/USD", "FLR/USD", "USDC/USD", "CBTC/USD",
	}
	
	// Mock base prices
	basePrices = map[string]float64{
		"ETH/USD":  2500.0,
		"BTC/USD":  45000.0,
		"FLR/USD":  0.05,
		"USDC/USD": 1.0,
		"CBTC/USD": 45000.0, // Same as BTC for Citrea
	}
)

func initializeFTSO() {
	log.Println("Initializing FTSO client...")
	
	// Initialize current prices with mock data
	for _, symbol := range supportedSymbols {
		price := basePrices[symbol]
		priceData := PriceData{
			Symbol:    symbol,
			Price:     price,
			Timestamp: time.Now().Unix(),
			Decimals:  8,
			Valid:     true,
		}
		
		pricesMutex.Lock()
		currentPrices[symbol] = priceData
		priceHistory[symbol] = []PriceData{priceData}
		pricesMutex.Unlock()
	}
	
	log.Println("FTSO client initialized with mock data")
}

func updatePriceFeeds() {
	pricesMutex.Lock()
	defer pricesMutex.Unlock()
	
	updated := 0
	for _, symbol := range supportedSymbols {
		basePrice := basePrices[symbol]
		
		// Add some random variation (±5%)
		variation := 0.05
		change := (rand.Float64() - 0.5) * 2 * variation
		newPrice := basePrice * (1 + change)
		
		priceData := PriceData{
			Symbol:    symbol,
			Price:     newPrice,
			Timestamp: time.Now().Unix(),
			Decimals:  8,
			Valid:     true,
		}
		
		currentPrices[symbol] = priceData
		
		// Keep last 100 price points
		history := priceHistory[symbol]
		history = append(history, priceData)
		if len(history) > 100 {
			history = history[1:]
		}
		priceHistory[symbol] = history
		
		updated++
	}
	
	if updated > 0 {
		log.Printf("Updated %d price feeds", updated)
	}
}

func handleGetPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	
	pricesMutex.RLock()
	priceData, exists := currentPrices[symbol]
	pricesMutex.RUnlock()
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Symbol not found"})
		return
	}
	
	// Check if price is stale (older than 2 minutes)
	if time.Now().Unix()-priceData.Timestamp > 120 {
		priceData.Valid = false
	}
	
	c.JSON(http.StatusOK, priceData)
}

func handleGetPriceHistory(c *gin.Context) {
	symbol := c.Param("symbol")
	limitStr := c.DefaultQuery("limit", "50")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	
	pricesMutex.RLock()
	history, exists := priceHistory[symbol]
	pricesMutex.RUnlock()
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Symbol not found"})
		return
	}
	
	// Return last 'limit' entries
	start := len(history) - limit
	if start < 0 {
		start = 0
	}
	
	response := PriceHistory{
		Symbol: symbol,
		Data:   history[start:],
	}
	
	c.JSON(http.StatusOK, response)
}

func handleUpdatePrice(c *gin.Context) {
	var request struct {
		Symbol string  `json:"symbol" binding:"required"`
		Price  float64 `json:"price" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Validate symbol
	validSymbol := false
	for _, s := range supportedSymbols {
		if s == request.Symbol {
			validSymbol = true
			break
		}
	}
	
	if !validSymbol {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported symbol"})
		return
	}
	
	if request.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price must be positive"})
		return
	}
	
	priceData := PriceData{
		Symbol:    request.Symbol,
		Price:     request.Price,
		Timestamp: time.Now().Unix(),
		Decimals:  8,
		Valid:     true,
	}
	
	pricesMutex.Lock()
	currentPrices[request.Symbol] = priceData
	
	// Add to history
	history := priceHistory[request.Symbol]
	history = append(history, priceData)
	if len(history) > 100 {
		history = history[1:]
	}
	priceHistory[request.Symbol] = history
	pricesMutex.Unlock()
	
	log.Printf("Price updated: %s = $%.2f", request.Symbol, request.Price)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    priceData,
	})
}

func handleGetSupportedSymbols(c *gin.Context) {
	pricesMutex.RLock()
	symbolsWithPrices := make(map[string]PriceData)
	for _, symbol := range supportedSymbols {
		if price, exists := currentPrices[symbol]; exists {
			symbolsWithPrices[symbol] = price
		}
	}
	pricesMutex.RUnlock()
	
	c.JSON(http.StatusOK, gin.H{
		"supported_symbols": supportedSymbols,
		"current_prices":    symbolsWithPrices,
		"total_count":       len(supportedSymbols),
	})
}

// Helper function to get price for contracts
func getPriceForPayment(symbol string) (PriceData, error) {
	pricesMutex.RLock()
	defer pricesMutex.RUnlock()
	
	priceData, exists := currentPrices[symbol]
	if !exists {
		return PriceData{}, fmt.Errorf("symbol not found: %s", symbol)
	}
	
	// Check if price is too stale
	if time.Now().Unix()-priceData.Timestamp > 300 { // 5 minutes
		return PriceData{}, fmt.Errorf("price too stale for %s", symbol)
	}
	
	return priceData, nil
}