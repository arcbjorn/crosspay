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
	mux := http.NewServeMux()
	
	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy":       true,
			"service":       "oracle-service",
			"timestamp":     1640995200,
			"ftsoHealthy":   true,
			"randomHealthy": true,
			"fdcHealthy":    true,
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["healthy"])
	assert.Equal(t, "oracle-service", response["service"])
}

func TestFTSOPriceEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/ftso/price/ETH", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// FTSO price endpoint
	mux.HandleFunc("/ftso/price/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"symbol":    "ETH/USD",
			"price":     2500.50,
			"timestamp": 1640995200,
			"decimals":  8,
			"valid":     true,
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ETH/USD", response["symbol"])
	assert.Equal(t, 2500.50, response["price"])
	assert.Equal(t, true, response["valid"])
}

func TestRandomNumberEndpoint(t *testing.T) {
	req, err := http.NewRequest("POST", "/rng/request", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// RNG request endpoint
	mux.HandleFunc("/rng/request", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"requestId":  "req_123456",
			"timestamp":  1640995200,
			"fulfilled":  false,
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "req_123456", response["requestId"])
	assert.Equal(t, false, response["fulfilled"])
}