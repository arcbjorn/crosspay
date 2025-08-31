package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PaymentData struct {
	ID           uint64    `json:"id"`
	Sender       string    `json:"sender"`
	SenderENS    string    `json:"sender_ens,omitempty"`
	Recipient    string    `json:"recipient"`
	RecipientENS string    `json:"recipient_ens,omitempty"`
	Token        string    `json:"token"`
	Amount       string    `json:"amount"`
	Fee          string    `json:"fee"`
	Status       string    `json:"status"`
	CreatedAt    int64     `json:"created_at"`
	CompletedAt  int64     `json:"completed_at,omitempty"`
	MetadataURI  string    `json:"metadata_uri"`
	TxHash       string    `json:"tx_hash"`
	ChainID      int       `json:"chain_id"`
	OraclePrice  string    `json:"oracle_price,omitempty"`
	RandomSeed   string    `json:"random_seed,omitempty"`
}

type Receipt struct {
	Payment     PaymentData       `json:"payment"`
	GeneratedAt time.Time         `json:"generated_at"`
	Version     string            `json:"version"`
	Format      string            `json:"format"`
	Signature   string            `json:"signature"`
	CID         string            `json:"cid,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

type GenerateReceiptRequest struct {
	PaymentID uint64                `json:"payment_id"`
	Format    string                `json:"format"` // "json" or "pdf"
	Language  string                `json:"language,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

type GenerateReceiptResponse struct {
	ReceiptID string    `json:"receipt_id"`
	CID       string    `json:"cid"`
	Format    string    `json:"format"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func handleGenerateReceipt(c *gin.Context) {
	var req GenerateReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Fetch payment data (mock implementation)
	paymentData, err := fetchPaymentData(req.PaymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Payment not found: %v", err)})
		return
	}

	// Generate receipt
	receipt, err := generateReceipt(paymentData, req.Format, req.Language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Receipt generation failed: %v", err)})
		return
	}

	// Convert receipt to uploadable format
	var uploadData []byte
	var filename string

	switch req.Format {
	case "pdf":
		uploadData, err = generatePDFReceipt(receipt)
		filename = fmt.Sprintf("receipt_%d.pdf", req.PaymentID)
	default:
		uploadData, err = json.MarshalIndent(receipt, "", "  ")
		filename = fmt.Sprintf("receipt_%d.json", req.PaymentID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Receipt formatting failed: %v", err)})
		return
	}

	// Upload to Filecoin
	cid, err := uploadToFilecoin(uploadData, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Storage upload failed: %v", err)})
		return
	}

	receipt.CID = cid

	response := GenerateReceiptResponse{
		ReceiptID: fmt.Sprintf("rcpt_%d_%d", req.PaymentID, time.Now().Unix()),
		CID:       cid,
		Format:    req.Format,
		Size:      int64(len(uploadData)),
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

func handleDownloadReceipt(c *gin.Context) {
	receiptID := c.Param("id")
	if receiptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Receipt ID required"})
		return
	}

	// Extract CID from receipt ID or lookup in database
	cid, err := getCIDFromReceiptID(receiptID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	// Retrieve from Filecoin
	data, metadata, err := retrieveFromFilecoin(cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Retrieval failed: %v", err)})
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", metadata["contentType"])
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", metadata["filename"]))
	c.Data(http.StatusOK, metadata["contentType"], data)
}

func handleVerifyReceipt(c *gin.Context) {
	cid := c.Param("cid")
	if cid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CID required"})
		return
	}

	// Retrieve and verify receipt
	data, _, err := retrieveFromFilecoin(cid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Receipt not found"})
		return
	}

	var receipt Receipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receipt format"})
		return
	}

	// Verify receipt signature and integrity
	isValid := verifyReceiptSignature(receipt)

	c.JSON(http.StatusOK, gin.H{
		"cid":       cid,
		"valid":     isValid,
		"payment_id": receipt.Payment.ID,
		"amount":    receipt.Payment.Amount,
		"status":    receipt.Payment.Status,
		"generated_at": receipt.GeneratedAt,
	})
}

func fetchPaymentData(paymentID uint64) (*PaymentData, error) {
	// Mock implementation - would fetch from blockchain
	log.Printf("Fetching payment data for ID: %d", paymentID)
	
	// Simulate API call delay
	time.Sleep(50 * time.Millisecond)
	
	return &PaymentData{
		ID:           paymentID,
		Sender:       "0x1234567890123456789012345678901234567890",
		SenderENS:    "alice.eth",
		Recipient:    "0x0987654321098765432109876543210987654321",
		RecipientENS: "bob.eth",
		Token:        "0x0000000000000000000000000000000000000000", // ETH
		Amount:       "1000000000000000000", // 1 ETH
		Fee:          "1000000000000000",    // 0.001 ETH
		Status:       "completed",
		CreatedAt:    time.Now().Unix() - 3600,
		CompletedAt:  time.Now().Unix() - 1800,
		MetadataURI:  "ipfs://QmTest123",
		TxHash:       "0xabcdef1234567890abcdef1234567890abcdef12",
		ChainID:      1135, // Lisk
		OraclePrice:  "2500.00", // ETH/USD
	}, nil
}

func generateReceipt(payment *PaymentData, format, language string) (*Receipt, error) {
	receipt := &Receipt{
		Payment:     *payment,
		GeneratedAt: time.Now(),
		Version:     "1.0",
		Format:      format,
		Metadata: map[string]string{
			"language":    language,
			"generator":   "crosspay-storage-worker",
			"network":     getNetworkName(payment.ChainID),
			"receipt_type": "payment_confirmation",
		},
	}

	// Generate signature for receipt integrity
	signature, err := signReceipt(receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to sign receipt: %v", err)
	}
	receipt.Signature = signature

	return receipt, nil
}

func generatePDFReceipt(receipt *Receipt) ([]byte, error) {
	// Mock PDF generation - would use actual PDF library
	log.Printf("Generating PDF receipt for payment %d", receipt.Payment.ID)
	
	// Simple mock PDF content
	pdfContent := fmt.Sprintf(`
CrossPay Payment Receipt
========================

Payment ID: %d
From: %s (%s)
To: %s (%s)
Amount: %s
Fee: %s
Status: %s
Created: %s
Completed: %s
Transaction: %s
Network: %s

Generated: %s
Signature: %s
`, 
		receipt.Payment.ID,
		receipt.Payment.SenderENS, receipt.Payment.Sender,
		receipt.Payment.RecipientENS, receipt.Payment.Recipient,
		receipt.Payment.Amount,
		receipt.Payment.Fee,
		receipt.Payment.Status,
		time.Unix(receipt.Payment.CreatedAt, 0).Format(time.RFC3339),
		time.Unix(receipt.Payment.CompletedAt, 0).Format(time.RFC3339),
		receipt.Payment.TxHash,
		getNetworkName(receipt.Payment.ChainID),
		receipt.GeneratedAt.Format(time.RFC3339),
		receipt.Signature,
	)

	return []byte(pdfContent), nil
}

func signReceipt(receipt *Receipt) (string, error) {
	// Mock signature generation
	data, err := json.Marshal(receipt.Payment)
	if err != nil {
		return "", err
	}
	
	// Would use actual cryptographic signing
	signature := fmt.Sprintf("sig_%x", len(data)+int(receipt.GeneratedAt.Unix()))
	return signature, nil
}

func verifyReceiptSignature(receipt Receipt) bool {
	// Mock signature verification
	expectedSig, err := signReceipt(&receipt)
	if err != nil {
		return false
	}
	return receipt.Signature == expectedSig
}

func getCIDFromReceiptID(receiptID string) (string, error) {
	// Mock CID lookup - would use database
	log.Printf("Looking up CID for receipt: %s", receiptID)
	return "bafybeigmock" + receiptID[5:], nil
}

func getNetworkName(chainID int) string {
	switch chainID {
	case 1135:
		return "Lisk"
	case 84532:
		return "Base Sepolia"
	case 5115:
		return "Citrea Testnet"
	default:
		return "Unknown"
	}
}

func uploadToFilecoin(data []byte, filename string) (string, error) {
	ctx := context.Background()
	result, err := storage.filecoinClient.Upload(ctx, data, filename, nil)
	if err != nil {
		return "", err
	}
	return result.CID, nil
}

func retrieveFromFilecoin(cid string) ([]byte, map[string]string, error) {
	ctx := context.Background()
	result, err := storage.filecoinClient.Retrieve(ctx, cid)
	if err != nil {
		return nil, nil, err
	}
	
	metadata := map[string]string{
		"filename":    result.Filename,
		"contentType": result.ContentType,
	}
	
	return result.Data, metadata, nil
}