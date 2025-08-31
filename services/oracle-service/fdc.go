package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ExternalProof struct {
	ID         string            `json:"id"`
	MerkleRoot string            `json:"merkle_root"`
	Proof      []string          `json:"proof"`
	Data       string            `json:"data"`
	DataHash   string            `json:"data_hash"`
	Timestamp  int64             `json:"timestamp"`
	Status     string            `json:"status"` // "submitted", "verified", "rejected"
	Metadata   map[string]string `json:"metadata"`
	VerifiedAt int64             `json:"verified_at,omitempty"`
}

type PaymentConfirmation struct {
	TxHash      string `json:"tx_hash"`
	BlockNumber int64  `json:"block_number"`
	ChainID     int    `json:"chain_id"`
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      string `json:"amount"`
	Token       string `json:"token"`
	Timestamp   int64  `json:"timestamp"`
}

var (
	externalProofs = make(map[string]*ExternalProof)
	proofsMutex    = sync.RWMutex{}
	proofCounter   = 0
)

func initializeFDC() {
	log.Println("Initializing FDC service...")
	// Mock initialization - would connect to Flare Data Connector
	log.Println("FDC service initialized")
}

func handleSubmitProof(c *gin.Context) {
	var request struct {
		MerkleRoot string            `json:"merkle_root" binding:"required"`
		Proof      []string          `json:"proof" binding:"required"`
		Data       string            `json:"data" binding:"required"`
		Metadata   map[string]string `json:"metadata"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Generate proof ID
	proofsMutex.Lock()
	proofCounter++
	proofID := fmt.Sprintf("fdc_%d_%d", time.Now().Unix(), proofCounter)
	
	// Calculate data hash
	dataHash := sha256.Sum256([]byte(request.Data))
	
	proof := &ExternalProof{
		ID:         proofID,
		MerkleRoot: request.MerkleRoot,
		Proof:      request.Proof,
		Data:       request.Data,
		DataHash:   hex.EncodeToString(dataHash[:]),
		Timestamp:  time.Now().Unix(),
		Status:     "submitted",
		Metadata:   request.Metadata,
	}
	
	if proof.Metadata == nil {
		proof.Metadata = make(map[string]string)
	}
	
	externalProofs[proofID] = proof
	proofsMutex.Unlock()
	
	log.Printf("External proof submitted: %s", proofID)
	
	c.JSON(http.StatusOK, gin.H{
		"proof_id":   proofID,
		"status":     "submitted",
		"data_hash":  proof.DataHash,
		"timestamp":  proof.Timestamp,
	})
}

func handleVerifyProof(c *gin.Context) {
	proofID := c.Param("proofId")
	
	proofsMutex.RLock()
	proof, exists := externalProofs[proofID]
	proofsMutex.RUnlock()
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proof not found"})
		return
	}
	
	// Perform Merkle proof verification
	isValid := verifyMerkleProof(proof.MerkleRoot, proof.Proof, proof.DataHash)
	
	response := gin.H{
		"proof_id":     proofID,
		"valid":        isValid,
		"merkle_root":  proof.MerkleRoot,
		"data_hash":    proof.DataHash,
		"timestamp":    proof.Timestamp,
		"status":       proof.Status,
		"verification_timestamp": time.Now().Unix(),
	}
	
	if !isValid {
		response["error"] = "Merkle proof verification failed"
	}
	
	c.JSON(http.StatusOK, response)
}

func handleConfirmProof(c *gin.Context) {
	var request struct {
		ProofID string `json:"proof_id" binding:"required"`
		Action  string `json:"action" binding:"required"` // "verify" or "reject"
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	if request.Action != "verify" && request.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'verify' or 'reject'"})
		return
	}
	
	proofsMutex.Lock()
	defer proofsMutex.Unlock()
	
	proof, exists := externalProofs[request.ProofID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proof not found"})
		return
	}
	
	if proof.Status != "submitted" {
		c.JSON(http.StatusConflict, gin.H{"error": "Proof already processed"})
		return
	}
	
	if request.Action == "verify" {
		// Verify the proof
		if verifyMerkleProof(proof.MerkleRoot, proof.Proof, proof.DataHash) {
			proof.Status = "verified"
		} else {
			proof.Status = "rejected"
			proof.Metadata["rejection_reason"] = "merkle_proof_invalid"
		}
	} else {
		proof.Status = "rejected"
		proof.Metadata["rejection_reason"] = "manual_rejection"
	}
	
	proof.VerifiedAt = time.Now().Unix()
	
	log.Printf("Proof %s %s", request.ProofID, proof.Status)
	
	c.JSON(http.StatusOK, gin.H{
		"proof_id":    request.ProofID,
		"status":      proof.Status,
		"verified_at": proof.VerifiedAt,
	})
}

func handlePaymentWebhook(c *gin.Context) {
	var confirmation PaymentConfirmation
	
	if err := c.ShouldBindJSON(&confirmation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment confirmation format"})
		return
	}
	
	// Validate required fields
	if confirmation.TxHash == "" || confirmation.From == "" || confirmation.To == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}
	
	// Process payment confirmation and create FDC proof
	proofID, err := createPaymentProof(confirmation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create proof: %v", err)})
		return
	}
	
	log.Printf("Payment confirmation received for tx %s, proof created: %s", confirmation.TxHash, proofID)
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "processed",
		"proof_id":  proofID,
		"tx_hash":   confirmation.TxHash,
		"timestamp": time.Now().Unix(),
	})
}

func createPaymentProof(confirmation PaymentConfirmation) (string, error) {
	// Create proof data from payment confirmation
	proofData := map[string]interface{}{
		"tx_hash":      confirmation.TxHash,
		"block_number": confirmation.BlockNumber,
		"chain_id":     confirmation.ChainID,
		"from":         confirmation.From,
		"to":           confirmation.To,
		"amount":       confirmation.Amount,
		"token":        confirmation.Token,
		"timestamp":    confirmation.Timestamp,
		"proof_type":   "payment_confirmation",
	}
	
	// Convert to JSON string
	dataBytes, err := json.Marshal(proofData)
	if err != nil {
		return "", err
	}
	
	// Generate mock Merkle proof
	dataHash := sha256.Sum256(dataBytes)
	merkleRoot := hex.EncodeToString(dataHash[:])
	
	// Create mock proof path
	proof := []string{
		hex.EncodeToString(sha256.New().Sum([]byte("proof1"))),
		hex.EncodeToString(sha256.New().Sum([]byte("proof2"))),
	}
	
	proofsMutex.Lock()
	proofCounter++
	proofID := fmt.Sprintf("payment_fdc_%d_%d", time.Now().Unix(), proofCounter)
	
	externalProof := &ExternalProof{
		ID:         proofID,
		MerkleRoot: merkleRoot,
		Proof:      proof,
		Data:       string(dataBytes),
		DataHash:   hex.EncodeToString(dataHash[:]),
		Timestamp:  time.Now().Unix(),
		Status:     "verified", // Auto-verify payment confirmations
		Metadata: map[string]string{
			"type":    "payment_confirmation",
			"tx_hash": confirmation.TxHash,
			"chain":   fmt.Sprintf("%d", confirmation.ChainID),
		},
		VerifiedAt: time.Now().Unix(),
	}
	
	externalProofs[proofID] = externalProof
	proofsMutex.Unlock()
	
	return proofID, nil
}

func verifyMerkleProof(merkleRoot string, proof []string, dataHash string) bool {
	// Simple mock verification - replace with actual Merkle proof verification
	if merkleRoot == "" || len(proof) == 0 || dataHash == "" {
		return false
	}
	
	// Mock verification logic
	computedHash := dataHash
	
	for _, proofElement := range proof {
		if proofElement == "" {
			return false
		}
		
		// Combine hashes (simplified)
		combined := computedHash + proofElement
		hash := sha256.Sum256([]byte(combined))
		computedHash = hex.EncodeToString(hash[:])
	}
	
	// In a real implementation, this would check if computedHash equals merkleRoot
	// For mock purposes, we'll just validate the format
	return len(merkleRoot) == 64 && len(dataHash) == 64
}

// Helper function to get all proofs for a specific transaction
func getProofsForTransaction(txHash string) []ExternalProof {
	proofsMutex.RLock()
	defer proofsMutex.RUnlock()
	
	var results []ExternalProof
	for _, proof := range externalProofs {
		if proof.Metadata["tx_hash"] == txHash {
			results = append(results, *proof)
		}
	}
	
	return results
}

// API endpoint to get proofs by transaction hash
func handleGetProofsByTx(c *gin.Context) {
	txHash := c.Query("tx_hash")
	if txHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tx_hash parameter required"})
		return
	}
	
	proofs := getProofsForTransaction(txHash)
	
	c.JSON(http.StatusOK, gin.H{
		"tx_hash": txHash,
		"proofs":  proofs,
		"count":   len(proofs),
	})
}