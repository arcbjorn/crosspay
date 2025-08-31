package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arcbjorn/crosspay/storage-worker/pkg/filecoin"
)

type StorageService struct {
	filecoinClient *filecoin.SynapseClient
}

type UploadRequest struct {
	Data        []byte            `json:"data"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"contentType"`
	Metadata    map[string]string `json:"metadata"`
}

type UploadResponse struct {
	CID       string    `json:"cid"`
	Size      int64     `json:"size"`
	Cost      string    `json:"cost"`
	Timestamp time.Time `json:"timestamp"`
}

type RetrieveResponse struct {
	Data        []byte            `json:"data"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"contentType"`
	Metadata    map[string]string `json:"metadata"`
	Size        int64             `json:"size"`
	Timestamp   time.Time         `json:"timestamp"`
}

type CostEstimate struct {
	SizeBytes    int64  `json:"size_bytes"`
	EstimatedFIL string `json:"estimated_fil"`
	USDEquiv     string `json:"usd_equivalent"`
}

var storage *StorageService

func initStorage() {
	apiURL := os.Getenv("SYNAPSE_API_URL")
	apiKey := os.Getenv("SYNAPSE_API_KEY")
	networkID := os.Getenv("FILECOIN_NETWORK")
	
	if apiKey == "" {
		log.Println("Warning: SYNAPSE_API_KEY not set, using mock mode")
		// In production, this should be an error
	}

	storage = &StorageService{
		filecoinClient: filecoin.NewSynapseClient(apiURL, apiKey, networkID),
	}
	
	log.Printf("Storage service initialized with Filecoin network: %s", networkID)
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "No file provided"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to read file"})
		return
	}

	// Upload to Filecoin via SynapseSDK
	ctx := context.Background()
	result, err := storage.filecoinClient.Upload(ctx, data, header.Filename, &filecoin.UploadOptions{
		DealDuration: 180, // 180 days
		PinToIPFS:    true,
		Metadata: map[string]string{
			"contentType": header.Header.Get("Content-Type"),
			"uploader":    r.RemoteAddr,
		},
	})
	if err != nil {
		log.Printf("Filecoin upload failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Upload failed: %v", err)})
		return
	}

	response := UploadResponse{
		CID:       result.CID,
		Size:      result.Size,
		Cost:      result.StorageCost,
		Timestamp: result.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleRetrieve(w http.ResponseWriter, r *http.Request) {
	// Extract CID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage/retrieve/")
	cid := strings.TrimSuffix(path, "/")
	if cid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "CID required"})
		return
	}

	// Retrieve from Filecoin via SynapseSDK
	ctx := context.Background()
	result, err := storage.filecoinClient.Retrieve(ctx, cid)
	if err != nil {
		log.Printf("Filecoin retrieval failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("Retrieval failed: %v", err)})
		return
	}

	response := RetrieveResponse{
		Data:        result.Data,
		Filename:    result.Filename,
		ContentType: result.ContentType,
		Metadata:    result.Metadata,
		Size:        result.Size,
		Timestamp:   result.RetrievedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleCostEstimate(w http.ResponseWriter, r *http.Request) {
	// Extract size from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage/cost/")
	sizeStr := strings.TrimSuffix(path, "/")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid size parameter"})
		return
	}

	// Get cost estimate from SynapseSDK
	ctx := context.Background()
	cost, err := storage.filecoinClient.EstimateStorageCost(ctx, size, 180) // 180 days
	if err != nil {
		log.Printf("Failed to get cost estimate: %v", err)
		// Fallback to calculation
		cost = calculateStorageCost(size)
	}
	
	response := CostEstimate{
		SizeBytes:    size,
		EstimatedFIL: cost,
		USDEquiv:     calculateUSDEquivalent(cost),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleListFiles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	files, err := storage.filecoinClient.ListFiles(ctx, 50, 0) // Default limit and offset
	if err != nil {
		log.Printf("Failed to list files: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to list files"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"files": files,
		"count": len(files),
	})
}

func handlePinToIPFS(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	// Extract CID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage/pin/")
	cid := strings.TrimSuffix(path, "/")
	if cid == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "CID required"})
		return
	}

	ctx := context.Background()
	err := storage.filecoinClient.PinToIPFS(ctx, cid)
	if err != nil {
		log.Printf("Failed to pin to IPFS: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to pin to IPFS"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Successfully pinned to IPFS",
		"cid":     cid,
	})
}

func handleDealStatus(w http.ResponseWriter, r *http.Request) {
	// Extract deal ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/storage/deal-status/")
	dealID := strings.TrimSuffix(path, "/")
	if dealID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Deal ID required"})
		return
	}

	ctx := context.Background()
	status, err := storage.filecoinClient.GetDealStatus(ctx, dealID)
	if err != nil {
		log.Printf("Failed to get deal status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Deal not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func handleNetworkInfo(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	info, err := storage.filecoinClient.GetNetworkInfo(ctx)
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to get network info"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

// Removed deprecated mock functions - now using SynapseClient directly

func calculateStorageCost(sizeBytes int64) string {
	// Fallback cost calculation - actual pricing comes from SynapseSDK
	// This is used only when SynapseSDK pricing is unavailable
	costFIL := float64(sizeBytes) * 0.000001 // 0.000001 FIL per byte (estimate)
	return fmt.Sprintf("%.6f", costFIL)
}

func calculateUSDEquivalent(filCost string) string {
	// TODO: Integrate with oracle service for FIL/USD price
	// For now, use approximate FIL price
	return "0.00"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}