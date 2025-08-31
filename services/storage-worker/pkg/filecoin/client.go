package filecoin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// SynapseClient wraps the SynapseSDK for Filecoin operations
type SynapseClient struct {
	apiURL    string
	apiKey    string
	client    *http.Client
	networkID string
}

// UploadOptions contains options for uploading files
type UploadOptions struct {
	DealDuration   int               `json:"deal_duration"`
	PinToIPFS      bool              `json:"pin_to_ipfs"`
	Metadata       map[string]string `json:"metadata"`
	Redundancy     int               `json:"redundancy"`
	StorageClass   string            `json:"storage_class"`
}

// UploadResult contains the result of a file upload
type UploadResult struct {
	CID         string            `json:"cid"`
	Size        int64             `json:"size"`
	DealID      string            `json:"deal_id"`
	StorageCost string            `json:"storage_cost"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
}

// RetrieveResult contains the result of a file retrieval
type RetrieveResult struct {
	Data        []byte            `json:"data"`
	CID         string            `json:"cid"`
	Filename    string            `json:"filename"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Metadata    map[string]string `json:"metadata"`
	RetrievedAt time.Time         `json:"retrieved_at"`
}

// DealStatus represents the status of a storage deal
type DealStatus struct {
	DealID      string    `json:"deal_id"`
	CID         string    `json:"cid"`
	Status      string    `json:"status"`
	StorageCost string    `json:"storage_cost"`
	Duration    int       `json:"duration"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// NewSynapseClient creates a new SynapseSDK client
func NewSynapseClient(apiURL, apiKey, networkID string) *SynapseClient {
	if apiURL == "" {
		apiURL = "https://api.synapse.org" // Default SynapseSDK API endpoint
	}
	if networkID == "" {
		networkID = "filecoin-calibration" // Default to Calibration testnet
	}

	return &SynapseClient{
		apiURL:    strings.TrimSuffix(apiURL, "/"),
		apiKey:    apiKey,
		networkID: networkID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Upload uploads data to Filecoin via SynapseSDK
func (c *SynapseClient) Upload(ctx context.Context, data []byte, filename string, options *UploadOptions) (*UploadResult, error) {
	if options == nil {
		options = &UploadOptions{
			DealDuration: 180, // 180 days default
			PinToIPFS:    true,
			Redundancy:   3,
			StorageClass: "standard",
			Metadata:     make(map[string]string),
		}
	}

	// Add filename to metadata
	options.Metadata["filename"] = filename
	options.Metadata["upload_time"] = time.Now().Format(time.RFC3339)

	// Prepare upload request
	uploadReq := map[string]interface{}{
		"data":          data,
		"network_id":    c.networkID,
		"deal_duration": options.DealDuration,
		"pin_to_ipfs":   options.PinToIPFS,
		"metadata":      options.Metadata,
		"redundancy":    options.Redundancy,
		"storage_class": options.StorageClass,
	}

	reqBody, err := json.Marshal(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal upload request: %w", err)
	}

	// Make API request
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/v1/storage/upload", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	log.Printf("Uploading file to Filecoin via SynapseSDK: %s (%d bytes)", filename, len(data))
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result UploadResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode upload response: %w", err)
	}

	log.Printf("File uploaded successfully: CID=%s, DealID=%s, Cost=%s", result.CID, result.DealID, result.StorageCost)
	
	return &result, nil
}

// Retrieve downloads data from Filecoin via SynapseSDK
func (c *SynapseClient) Retrieve(ctx context.Context, cid string) (*RetrieveResult, error) {
	if cid == "" {
		return nil, errors.New("CID cannot be empty")
	}

	log.Printf("Retrieving file from Filecoin via SynapseSDK: CID=%s", cid)

	// Make API request
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL+"/v1/storage/retrieve/"+cid, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("X-Network-ID", c.networkID)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make retrieve request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("file not found: CID=%s", cid)
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("retrieval failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response (assuming JSON format with metadata)
	var result RetrieveResult
	if err := json.Unmarshal(data, &result); err != nil {
		// If it's not JSON, treat as raw data
		result = RetrieveResult{
			Data:        data,
			CID:         cid,
			Size:        int64(len(data)),
			RetrievedAt: time.Now(),
		}
	}

	log.Printf("File retrieved successfully: CID=%s, Size=%d bytes", cid, len(result.Data))
	
	return &result, nil
}

// GetDealStatus retrieves the status of a storage deal
func (c *SynapseClient) GetDealStatus(ctx context.Context, dealID string) (*DealStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL+"/v1/storage/deal/"+dealID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make deal status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deal status request failed with status %d", resp.StatusCode)
	}

	var status DealStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode deal status response: %w", err)
	}

	return &status, nil
}

// EstimateStorageCost estimates the cost of storing data
func (c *SynapseClient) EstimateStorageCost(ctx context.Context, sizeBytes int64, duration int) (string, error) {
	reqData := map[string]interface{}{
		"size_bytes": sizeBytes,
		"duration":   duration,
		"network_id": c.networkID,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cost estimate request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/v1/storage/cost-estimate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make cost estimate request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("cost estimate request failed with status %d", resp.StatusCode)
	}

	var result struct {
		EstimatedCost string `json:"estimated_cost"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode cost estimate response: %w", err)
	}

	return result.EstimatedCost, nil
}

// ListFiles lists files stored for the current API key
func (c *SynapseClient) ListFiles(ctx context.Context, limit int, offset int) ([]UploadResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", 
		fmt.Sprintf("%s/v1/storage/files?limit=%d&offset=%d", c.apiURL, limit, offset), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make list files request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list files request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Files []UploadResult `json:"files"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode list files response: %w", err)
	}

	return result.Files, nil
}

// PinToIPFS pins a file to IPFS
func (c *SynapseClient) PinToIPFS(ctx context.Context, cid string) error {
	reqData := map[string]interface{}{
		"cid": cid,
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("failed to marshal pin request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/v1/ipfs/pin", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make pin request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("pin request failed with status %d", resp.StatusCode)
	}

	log.Printf("File pinned to IPFS successfully: CID=%s", cid)
	return nil
}

// GetNetworkInfo returns information about the Filecoin network
func (c *SynapseClient) GetNetworkInfo(ctx context.Context) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL+"/v1/network/info", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make network info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network info request failed with status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode network info response: %w", err)
	}

	return result, nil
}