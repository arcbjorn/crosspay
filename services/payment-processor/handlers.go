package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Service clients (would be properly initialized with HTTP clients)
var (
	storageServiceURL = "http://storage-worker:8080"
	oracleServiceURL  = "http://oracle-service:8081" 
	ensServiceURL     = "http://ens-resolver:8082"
)

// Payment handlers
func handleCreatePayment(c *gin.Context) {
	var request struct {
		Recipient    string `json:"recipient" binding:"required"`
		Token        string `json:"token" binding:"required"`
		Amount       string `json:"amount" binding:"required"`
		MetadataURI  string `json:"metadata_uri"`
		SenderENS    string `json:"sender_ens"`
		RecipientENS string `json:"recipient_ens"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Resolve ENS names if provided
	if request.SenderENS != "" {
		resolvedSender, err := resolveENSName(request.SenderENS)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to resolve sender ENS: %v", err)})
			return
		}
		log.Printf("Resolved sender ENS %s to %s", request.SenderENS, resolvedSender)
	}

	if request.RecipientENS != "" {
		resolvedRecipient, err := resolveENSName(request.RecipientENS)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to resolve recipient ENS: %v", err)})
			return
		}
		if resolvedRecipient != request.Recipient {
			log.Printf("Warning: Recipient address %s doesn't match resolved ENS %s -> %s", 
				request.Recipient, request.RecipientENS, resolvedRecipient)
		}
	}

	// Get current price from oracle
	oraclePrice, err := getOraclePrice("ETH/USD")
	if err != nil {
		log.Printf("Warning: Failed to get oracle price: %v", err)
		oraclePrice = "0"
	}

	// Mock payment creation (would interact with blockchain)
	paymentID := time.Now().Unix()
	
	// Generate receipt automatically
	receiptCID, err := generatePaymentReceipt(paymentID, request)
	if err != nil {
		log.Printf("Warning: Failed to generate receipt: %v", err)
	}

	response := gin.H{
		"payment_id":     paymentID,
		"status":         "pending",
		"oracle_price":   oraclePrice,
		"receipt_cid":    receiptCID,
		"created_at":     time.Now().Unix(),
		"tx_hash":        fmt.Sprintf("0x%x", paymentID), // Mock tx hash
	}

	c.JSON(http.StatusCreated, response)
}

func handleCompletePayment(c *gin.Context) {
	paymentID := c.Param("id")
	
	// Mock payment completion
	log.Printf("Completing payment: %s", paymentID)
	
	c.JSON(http.StatusOK, gin.H{
		"payment_id":   paymentID,
		"status":       "completed",
		"completed_at": time.Now().Unix(),
	})
}

func handleRefundPayment(c *gin.Context) {
	paymentID := c.Param("id")
	
	// Mock payment refund
	log.Printf("Refunding payment: %s", paymentID)
	
	c.JSON(http.StatusOK, gin.H{
		"payment_id": paymentID,
		"status":     "refunded",
		"refunded_at": time.Now().Unix(),
	})
}

func handleGetPayment(c *gin.Context) {
	paymentID := c.Param("id")
	
	// Mock payment retrieval
	c.JSON(http.StatusOK, gin.H{
		"payment_id":    paymentID,
		"sender":        "0x1234...",
		"recipient":     "0x5678...",
		"amount":        "1000000000000000000",
		"status":        "completed",
		"created_at":    time.Now().Unix() - 3600,
		"completed_at":  time.Now().Unix() - 1800,
	})
}

func handleGetUserPayments(c *gin.Context) {
	address := c.Param("address")
	
	// Mock user payments
	payments := []gin.H{
		{
			"payment_id": 1,
			"recipient":  "0x9999...",
			"amount":     "500000000000000000",
			"status":     "completed",
			"created_at": time.Now().Unix() - 7200,
		},
		{
			"payment_id": 2,
			"sender":     "0x8888...",
			"amount":     "750000000000000000",
			"status":     "pending",
			"created_at": time.Now().Unix() - 1800,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"address":  address,
		"payments": payments,
		"count":    len(payments),
	})
}

// Receipt handlers
func handleGenerateReceipt(c *gin.Context) {
	paymentID := c.Param("paymentId")
	
	var request struct {
		Format   string `json:"format"`
		Language string `json:"language"`
	}
	
	c.ShouldBindJSON(&request)
	
	if request.Format == "" {
		request.Format = "json"
	}
	if request.Language == "" {
		request.Language = "en"
	}
	
	// Call storage worker to generate receipt
	receiptData := map[string]interface{}{
		"payment_id": paymentID,
		"format":     request.Format,
		"language":   request.Language,
	}
	
	resp, err := makeServiceCall("POST", storageServiceURL+"/api/receipts/generate", receiptData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate receipt: %v", err)})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleDownloadReceipt(c *gin.Context) {
	receiptID := c.Param("id")
	
	// Proxy to storage worker
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/receipts/download/"+receiptID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to download receipt: %v", err)})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleVerifyReceipt(c *gin.Context) {
	cid := c.Param("cid")
	
	// Proxy to storage worker
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/receipts/verify/"+cid, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to verify receipt: %v", err)})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleGetReceiptsByPayment(c *gin.Context) {
	paymentID := c.Param("paymentId")
	
	// Mock receipts for payment
	c.JSON(http.StatusOK, gin.H{
		"payment_id": paymentID,
		"receipts": []gin.H{
			{
				"receipt_id": "rcpt_1",
				"cid":        "bafybei...",
				"format":     "json",
				"language":   "en",
				"created_at": time.Now().Unix(),
			},
		},
	})
}

// Oracle handlers
func handleGetPrice(c *gin.Context) {
	symbol := c.Param("symbol")
	
	price, err := getOraclePrice(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"price":  price,
		"timestamp": time.Now().Unix(),
	})
}

func handleRequestRandom(c *gin.Context) {
	resp, err := makeServiceCall("POST", oracleServiceURL+"/api/random/request", map[string]string{
		"requester": "payment-processor",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleRandomStatus(c *gin.Context) {
	requestID := c.Param("requestId")
	
	resp, err := makeServiceCall("GET", oracleServiceURL+"/api/random/status/"+requestID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleSubmitProof(c *gin.Context) {
	var proofData map[string]interface{}
	if err := c.ShouldBindJSON(&proofData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proof data"})
		return
	}
	
	resp, err := makeServiceCall("POST", oracleServiceURL+"/api/fdc/proof/submit", proofData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleVerifyProof(c *gin.Context) {
	proofID := c.Param("proofId")
	
	resp, err := makeServiceCall("GET", oracleServiceURL+"/api/fdc/proof/verify/"+proofID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// ENS handlers
func handleResolveName(c *gin.Context) {
	name := c.Param("name")
	
	address, err := resolveENSName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"name":    name,
		"address": address,
	})
}

func handleReverseResolve(c *gin.Context) {
	address := c.Param("address")
	
	resp, err := makeServiceCall("GET", ensServiceURL+"/api/ens/reverse/"+address, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleBatchResolve(c *gin.Context) {
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	
	resp, err := makeServiceCall("POST", ensServiceURL+"/api/ens/resolve/batch", request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// Storage handlers  
func handleUploadFile(c *gin.Context) {
	resp, err := makeServiceCall("POST", storageServiceURL+"/api/storage/upload", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleRetrieveFile(c *gin.Context) {
	cid := c.Param("cid")
	
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/storage/retrieve/"+cid, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

func handleEstimateCost(c *gin.Context) {
	size := c.Param("size")
	
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/storage/cost/"+size, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, resp)
}

// Analytics handlers
func handleGetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_payments":    1000,
		"completed_payments": 850,
		"total_volume":      "1250000000000000000000", // 1250 ETH
		"receipts_generated": 750,
		"receipts_verified":  600,
		"oracle_requests":    500,
		"ens_resolutions":    300,
	})
}

func handleGetPaymentVolume(c *gin.Context) {
	// Mock volume data
	c.JSON(http.StatusOK, gin.H{
		"daily_volume": []gin.H{
			{"date": "2024-01-01", "volume": "50000000000000000000"},
			{"date": "2024-01-02", "volume": "75000000000000000000"},
			{"date": "2024-01-03", "volume": "100000000000000000000"},
		},
		"total_volume": "1250000000000000000000",
	})
}

func handleGetReceiptStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_receipts":    750,
		"verified_receipts": 600,
		"by_format": gin.H{
			"json": 450,
			"pdf":  300,
		},
		"by_language": gin.H{
			"en": 500,
			"es": 150,
			"fr": 100,
		},
	})
}

// Utility functions
func makeServiceCall(method, url string, data interface{}) (map[string]interface{}, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result, nil
}

func getOraclePrice(symbol string) (string, error) {
	resp, err := makeServiceCall("GET", oracleServiceURL+"/api/ftso/price/"+symbol, nil)
	if err != nil {
		return "", err
	}
	
	if price, ok := resp["price"].(float64); ok {
		return strconv.FormatFloat(price, 'f', 2, 64), nil
	}
	
	return "0", fmt.Errorf("invalid price format")
}

func resolveENSName(name string) (string, error) {
	resp, err := makeServiceCall("GET", ensServiceURL+"/api/ens/resolve/"+name, nil)
	if err != nil {
		return "", err
	}
	
	if address, ok := resp["address"].(string); ok {
		return address, nil
	}
	
	return "", fmt.Errorf("invalid address format")
}

func generatePaymentReceipt(paymentID int64, request interface{}) (string, error) {
	receiptData := map[string]interface{}{
		"payment_id": paymentID,
		"format":     "json",
		"language":   "en",
	}
	
	resp, err := makeServiceCall("POST", storageServiceURL+"/api/receipts/generate", receiptData)
	if err != nil {
		return "", err
	}
	
	if cid, ok := resp["cid"].(string); ok {
		return cid, nil
	}
	
	return "", fmt.Errorf("failed to get CID from response")
}