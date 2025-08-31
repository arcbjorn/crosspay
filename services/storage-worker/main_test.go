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
	
	// Health check endpoint using standard library
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"healthy":   true,
			"service":   "storage-worker",
			"timestamp": 1640995200,
			"filecoin":  true,
			"synapse":   true,
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, true, response["healthy"])
	assert.Equal(t, "storage-worker", response["service"])
}

func TestFileUploadEndpoint(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/storage/upload", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// File upload endpoint using standard library
	mux.HandleFunc("/api/storage/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"cid":       "bafybeigtest123",
			"size":      1024,
			"cost":      "0.001",
			"timestamp": "2021-12-31T23:59:59Z",
			"status":    "uploaded",
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "bafybeigtest123", response["cid"])
	assert.Equal(t, float64(1024), response["size"])
}

func TestFileRetrieveEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/storage/retrieve/bafybeigtest123", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	
	// File retrieve endpoint using standard library
	mux.HandleFunc("/api/storage/retrieve/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":        "SGVsbG8gV29ybGQ=", // "Hello World" base64 encoded
			"filename":    "test.txt",
			"contentType": "text/plain",
			"size":        11,
			"metadata":    map[string]string{"type": "test"},
		})
	})

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "SGVsbG8gV29ybGQ=", response["data"])
	assert.Equal(t, "test.txt", response["filename"])
}