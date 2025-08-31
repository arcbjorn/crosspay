package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Service clients (would be properly initialized with HTTP clients)
var (
	storageServiceURL = "http://storage-worker:8080"
	oracleServiceURL  = "http://oracle-service:8081" 
	ensServiceURL     = "http://ens-resolver:8082"
)

// Payment handlers
func handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var request struct {
		Recipient    string `json:"recipient"`
		Token        string `json:"token"`
		Amount       string `json:"amount"`
		MetadataURI  string `json:"metadata_uri"`
		SenderENS    string `json:"sender_ens"`
		RecipientENS string `json:"recipient_ens"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid request format"})
		return
	}

	if request.Recipient == "" || request.Token == "" || request.Amount == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Missing required fields"})
		return
	}

	// Resolve ENS names if provided
	if request.SenderENS != "" {
		resolvedSender, err := resolveENSName(request.SenderENS)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Failed to resolve sender ENS: %v", err)})
			return
		}
		log.Printf("Resolved sender ENS %s to %s", request.SenderENS, resolvedSender)
	}

	if request.RecipientENS != "" {
		resolvedRecipient, err := resolveENSName(request.RecipientENS)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Failed to resolve recipient ENS: %v", err)})
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

	response := map[string]interface{}{
		"payment_id":     paymentID,
		"status":         "pending",
		"oracle_price":   oraclePrice,
		"receipt_cid":    receiptCID,
		"created_at":     time.Now().Unix(),
		"tx_hash":        fmt.Sprintf("0x%x", paymentID), // Mock tx hash
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func handleCompletePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/payments/complete/")
	paymentID := strings.TrimSuffix(path, "/")
	
	// Mock payment completion
	log.Printf("Completing payment: %s", paymentID)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_id":   paymentID,
		"status":       "completed",
		"completed_at": time.Now().Unix(),
	})
}

func handleRefundPayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/payments/refund/")
	paymentID := strings.TrimSuffix(path, "/")
	
	// Mock payment refund
	log.Printf("Refunding payment: %s", paymentID)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_id": paymentID,
		"status":     "refunded",
		"refunded_at": time.Now().Unix(),
	})
}

func handleGetPayment(w http.ResponseWriter, r *http.Request) {
	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/payments/")
	paymentID := strings.TrimSuffix(path, "/")
	
	// Mock payment retrieval
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_id":    paymentID,
		"sender":        "0x1234...",
		"recipient":     "0x5678...",
		"amount":        "1000000000000000000",
		"status":        "completed",
		"created_at":    time.Now().Unix() - 3600,
		"completed_at":  time.Now().Unix() - 1800,
	})
}

func handleGetUserPayments(w http.ResponseWriter, r *http.Request) {
	// Extract address from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/payments/user/")
	address := strings.TrimSuffix(path, "/")
	
	// Mock user payments
	payments := []map[string]interface{}{
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
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address":  address,
		"payments": payments,
		"count":    len(payments),
	})
}

// Receipt handlers
func handleGenerateReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/receipts/generate/")
	paymentID := strings.TrimSuffix(path, "/")
	
	var request struct {
		Format   string `json:"format"`
		Language string `json:"language"`
	}
	
	json.NewDecoder(r.Body).Decode(&request)
	
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Failed to generate receipt: %v", err)})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleDownloadReceipt(w http.ResponseWriter, r *http.Request) {
	// Extract receipt ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/receipts/download/")
	receiptID := strings.TrimSuffix(path, "/")
	
	// Proxy to storage worker
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/receipts/download/"+receiptID, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Failed to download receipt: %v", err)})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleVerifyReceipt(w http.ResponseWriter, r *http.Request) {
	// Extract CID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/receipts/verify/")
	cid := strings.TrimSuffix(path, "/")
	
	// Proxy to storage worker
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/receipts/verify/"+cid, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Failed to verify receipt: %v", err)})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleGetReceiptsByPayment(w http.ResponseWriter, r *http.Request) {
	// Extract payment ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/receipts/payment/")
	paymentID := strings.TrimSuffix(path, "/")
	
	// Mock receipts for payment
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_id": paymentID,
		"receipts": []map[string]interface{}{
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
func handleGetPrice(w http.ResponseWriter, r *http.Request) {
	// Extract symbol from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/oracle/price/")
	symbol := strings.TrimSuffix(path, "/")
	
	price, err := getOraclePrice(symbol)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol": symbol,
		"price":  price,
		"timestamp": time.Now().Unix(),
	})
}

func handleRequestRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	resp, err := makeServiceCall("POST", oracleServiceURL+"/api/random/request", map[string]string{
		"requester": "payment-processor",
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleRandomStatus(w http.ResponseWriter, r *http.Request) {
	// Extract request ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/oracle/random/status/")
	requestID := strings.TrimSuffix(path, "/")
	
	resp, err := makeServiceCall("GET", oracleServiceURL+"/api/random/status/"+requestID, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleSubmitProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var proofData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&proofData); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid proof data"})
		return
	}
	
	resp, err := makeServiceCall("POST", oracleServiceURL+"/api/fdc/proof/submit", proofData)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleVerifyProof(w http.ResponseWriter, r *http.Request) {
	// Extract proof ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/oracle/proof/verify/")
	proofID := strings.TrimSuffix(path, "/")
	
	resp, err := makeServiceCall("GET", oracleServiceURL+"/api/fdc/proof/verify/"+proofID, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ENS handlers
func handleResolveName(w http.ResponseWriter, r *http.Request) {
	// Extract name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/ens/resolve/")
	name := strings.TrimSuffix(path, "/")
	
	address, err := resolveENSName(name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":    name,
		"address": address,
	})
}

func handleReverseResolve(w http.ResponseWriter, r *http.Request) {
	// Extract address from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/ens/reverse/")
	address := strings.TrimSuffix(path, "/")
	
	resp, err := makeServiceCall("GET", ensServiceURL+"/api/ens/reverse/"+address, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleBatchResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid request"})
		return
	}
	
	resp, err := makeServiceCall("POST", ensServiceURL+"/api/ens/resolve/batch", request)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Storage handlers  
func handleUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	resp, err := makeServiceCall("POST", storageServiceURL+"/api/storage/upload", nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleRetrieveFile(w http.ResponseWriter, r *http.Request) {
	// Extract CID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage/retrieve/")
	cid := strings.TrimSuffix(path, "/")
	
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/storage/retrieve/"+cid, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleEstimateCost(w http.ResponseWriter, r *http.Request) {
	// Extract size from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage/cost/")
	size := strings.TrimSuffix(path, "/")
	
	resp, err := makeServiceCall("GET", storageServiceURL+"/api/storage/cost/"+size, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Analytics handlers
func handleGetStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_payments":    1000,
		"completed_payments": 850,
		"total_volume":      "1250000000000000000000", // 1250 ETH
		"receipts_generated": 750,
		"receipts_verified":  600,
		"oracle_requests":    500,
		"ens_resolutions":    300,
	})
}

func handleGetPaymentVolume(w http.ResponseWriter, r *http.Request) {
	// Mock volume data
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"daily_volume": []map[string]interface{}{
			{"date": "2024-01-01", "volume": "50000000000000000000"},
			{"date": "2024-01-02", "volume": "75000000000000000000"},
			{"date": "2024-01-03", "volume": "100000000000000000000"},
		},
		"total_volume": "1250000000000000000000",
	})
}

func handleGetReceiptStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_receipts":    750,
		"verified_receipts": 600,
		"by_format": map[string]interface{}{
			"json": 450,
			"pdf":  300,
		},
		"by_language": map[string]interface{}{
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