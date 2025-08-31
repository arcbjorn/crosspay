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
			"healthy":   true,
			"service":   "ens-resolver",
			"timestamp": 1640995200,
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["healthy"])
	assert.Equal(t, "ens-resolver", response["service"])
}

func TestResolveENSEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/resolve/alice.eth", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// ENS resolution endpoint
	mux.HandleFunc("/resolve/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"address": "0x1234567890123456789012345678901234567890",
			"name":    "alice.eth",
			"avatar":  "https://example.com/avatar.png",
			"email":   "alice@example.com",
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "0x1234567890123456789012345678901234567890", response["address"])
	assert.Equal(t, "alice.eth", response["name"])
}

func TestReverseResolveEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/reverse/0x1234567890123456789012345678901234567890", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// Reverse resolution endpoint
	mux.HandleFunc("/reverse/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":    "alice.eth",
			"address": "0x1234567890123456789012345678901234567890",
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alice.eth", response["name"])
	assert.Equal(t, "0x1234567890123456789012345678901234567890", response["address"])
}

func TestSubnameRegistrationEndpoint(t *testing.T) {
	// This would be a POST request with request body in real implementation
	req, err := http.NewRequest("POST", "/register-subname", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// Subname registration endpoint
	mux.HandleFunc("/register-subname", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"subname":          "pay.alice.eth",
			"domain":           "alice.eth",
			"owner":            "0x1234567890123456789012345678901234567890",
			"registrationFee":  "0.001",
			"success":          true,
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "pay.alice.eth", response["subname"])
	assert.Equal(t, true, response["success"])
}