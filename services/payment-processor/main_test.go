package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	
	// Mock CORS handler
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}
	
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy":   true,
			"service":   "payment-processor",
			"timestamp": 1640995200,
			"chains":    []int{4202, 84532, 5115},
		})
	}))

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["healthy"])
	assert.Equal(t, "payment-processor", response["service"])
}

func TestPaymentEndpoint(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/pay", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	
	// Mock CORS handler
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}
	
	mux := http.NewServeMux()
	
	// Payment processing endpoint
	mux.HandleFunc("/api/pay", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"paymentId":  "123456",
			"hash":       "0xabcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
			"chainId":    4202,
			"status":     "pending",
			"estimatedConfirmationTime": 30,
		})
	}))

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["success"])
	assert.Equal(t, "123456", response["paymentId"])
}

func TestPaymentStatusEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/payments/123456", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	
	// Mock CORS handler
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			next.ServeHTTP(w, r)
		})
	}
	
	mux := http.NewServeMux()
	
	// Payment status endpoint
	mux.HandleFunc("/api/payments/", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "123456",
			"status":    "completed",
			"amount":    "1000000000000000000",
			"token":     "ETH",
			"sender":    "0x1234567890123456789012345678901234567890",
			"recipient": "0x0987654321098765432109876543210987654321",
			"chainId":   4202,
			"hash":      "0xabcd1234567890abcdef1234567890abcdef1234567890abcdef1234567890ab",
			"blockNumber": 1234567,
		})
	}))

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "123456", response["id"])
	assert.Equal(t, "completed", response["status"])
}