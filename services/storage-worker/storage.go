package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type StorageService struct {
	// SynapseSDK client would be initialized here
	// client *synapsesdk.Client
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

var storage = &StorageService{}

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

	// Mock SynapseSDK upload - replace with actual implementation
	cid, err := uploadToFilecoin(data, header.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Upload failed: %v", err)})
		return
	}

	cost := calculateStorageCost(int64(len(data)))

	response := UploadResponse{
		CID:       cid,
		Size:      int64(len(data)),
		Cost:      cost,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func handleRetrieve(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CID required"})
		return
	}

	// Mock SynapseSDK retrieval - replace with actual implementation
	data, metadata, err := retrieveFromFilecoin(cid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Retrieval failed: %v", err)})
		return
	}

	response := RetrieveResponse{
		Data:        data,
		Filename:    metadata["filename"],
		ContentType: metadata["contentType"],
		Metadata:    metadata,
		Size:        int64(len(data)),
		Timestamp:   time.Now(),
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

	cost := calculateStorageCost(size)
	
	response := CostEstimate{
		SizeBytes:    size,
		EstimatedFIL: cost,
		USDEquiv:     calculateUSDEquivalent(cost),
	}

	c.JSON(http.StatusOK, response)
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