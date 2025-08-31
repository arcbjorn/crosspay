package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"./pkg/filecoin"
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

func handleUpload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Upload to Filecoin via SynapseSDK
	ctx := context.Background()
	result, err := storage.filecoinClient.Upload(ctx, data, header.Filename, &filecoin.UploadOptions{
		DealDuration: 180, // 180 days
		PinToIPFS:    true,
		Metadata: map[string]string{
			"contentType": header.Header.Get("Content-Type"),
			"uploader":    c.ClientIP(),
		},
	})
	if err != nil {
		log.Printf("Filecoin upload failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Upload failed: %v", err)})
		return
	}

	response := UploadResponse{
		CID:       result.CID,
		Size:      result.Size,
		Cost:      result.StorageCost,
		Timestamp: result.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}

func handleRetrieve(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CID required"})
		return
	}

	// Retrieve from Filecoin via SynapseSDK
	ctx := context.Background()
	result, err := storage.filecoinClient.Retrieve(ctx, cid)
	if err != nil {
		log.Printf("Filecoin retrieval failed: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Retrieval failed: %v", err)})
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

	c.JSON(http.StatusOK, response)
}

func handleCostEstimate(c *gin.Context) {
	sizeStr := c.Param("size")
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid size parameter"})
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

	c.JSON(http.StatusOK, response)
}

func handleListFiles(c *gin.Context) {
	ctx := context.Background()
	files, err := storage.filecoinClient.ListFiles(ctx, 50, 0) // Default limit and offset
	if err != nil {
		log.Printf("Failed to list files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"count": len(files),
	})
}

func handlePinToIPFS(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CID required"})
		return
	}

	ctx := context.Background()
	err := storage.filecoinClient.PinToIPFS(ctx, cid)
	if err != nil {
		log.Printf("Failed to pin to IPFS: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pin to IPFS"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully pinned to IPFS",
		"cid":     cid,
	})
}

func handleDealStatus(c *gin.Context) {
	dealID := c.Param("dealId")
	if dealID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deal ID required"})
		return
	}

	ctx := context.Background()
	status, err := storage.filecoinClient.GetDealStatus(ctx, dealID)
	if err != nil {
		log.Printf("Failed to get deal status: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Deal not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

func handleNetworkInfo(c *gin.Context) {
	ctx := context.Background()
	info, err := storage.filecoinClient.GetNetworkInfo(ctx)
	if err != nil {
		log.Printf("Failed to get network info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get network info"})
		return
	}

	c.JSON(http.StatusOK, info)
}

func uploadToFilecoin(data []byte, filename string) (string, error) {
	// Mock implementation - replace with SynapseSDK
	// This would use the actual SynapseSDK client
	
	log.Printf("Uploading file: %s, size: %d bytes", filename, len(data))
	
	// Simulate upload delay
	time.Sleep(100 * time.Millisecond)
	
	// Generate mock CID (would be returned by SynapseSDK)
	cid := fmt.Sprintf("bafybeig%s%d", filename[:min(len(filename), 8)], time.Now().Unix())
	
	log.Printf("File uploaded successfully, CID: %s", cid)
	return cid, nil
}

func retrieveFromFilecoin(cid string) ([]byte, map[string]string, error) {
	// Mock implementation - replace with SynapseSDK
	log.Printf("Retrieving file with CID: %s", cid)
	
	// Simulate retrieval delay
	time.Sleep(50 * time.Millisecond)
	
	// Mock data
	data := []byte(fmt.Sprintf("Mock file content for CID: %s", cid))
	metadata := map[string]string{
		"filename":    "receipt.json",
		"contentType": "application/json",
		"uploadTime":  time.Now().Format(time.RFC3339),
	}
	
	return data, metadata, nil
}

func calculateStorageCost(sizeBytes int64) string {
	// Mock cost calculation - replace with actual Filecoin pricing
	costFIL := float64(sizeBytes) * 0.000001 // 0.000001 FIL per byte
	return fmt.Sprintf("%.6f", costFIL)
}

func calculateUSDEquivalent(filCost string) string {
	// Mock USD conversion - would use oracle price
	return "0.00"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}