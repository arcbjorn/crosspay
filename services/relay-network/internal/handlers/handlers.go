package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/crosspay/relay-network/internal/p2p"
)

type Handler struct {
	validator ValidatorNode
	network   P2PNetwork
}

type ValidatorNode interface {
	GetAddress() string
	GetStatus() string
	GetStake() string
	IsRegistered() bool
	GetPendingValidationCount() int
	ProcessValidationRequest(req *ValidationRequest) error
	GetValidationStatus(requestID uint64) (*ValidationRequest, bool)
	GetSignatures(requestID uint64) map[string]string
}

type P2PNetwork interface {
	GetPeers() []*p2p.Peer
	GetPeerCount() int
	IsRunning() bool
	BroadcastValidationRequest(req *p2p.ValidationMessage) error
	BroadcastSignature(requestID uint64, signature string) error
}

type ValidationRequest struct {
	ID           uint64    `json:"id"`
	PaymentID    uint64    `json:"payment_id"`
	MessageHash  string    `json:"message_hash"`
	RequiredSigs int       `json:"required_signatures"`
	Deadline     time.Time `json:"deadline"`
	IsHighValue  bool      `json:"is_high_value"`
}

type HealthResponse struct {
	Status              string    `json:"status"`
	Timestamp           time.Time `json:"timestamp"`
	ValidatorAddress    string    `json:"validator_address"`
	IsRegistered        bool      `json:"is_registered"`
	Stake               string    `json:"stake"`
	PeerCount           int       `json:"peer_count"`
	PendingValidations  int       `json:"pending_validations"`
	NetworkRunning      bool      `json:"network_running"`
}

type StatusResponse struct {
	ValidatorAddress   string                 `json:"validator_address"`
	Status             string                 `json:"status"`
	IsRegistered       bool                   `json:"is_registered"`
	Stake              string                 `json:"stake"`
	PeerCount          int                    `json:"peer_count"`
	PendingValidations int                    `json:"pending_validations"`
	NetworkRunning     bool                   `json:"network_running"`
	Peers              []*p2p.Peer            `json:"peers"`
}

type ValidationRequestPayload struct {
	PaymentID    uint64 `json:"payment_id"`
	MessageHash  string `json:"message_hash"`
	RequiredSigs int    `json:"required_signatures"`
	IsHighValue  bool   `json:"is_high_value"`
}

type SignMessagePayload struct {
	RequestID   uint64 `json:"request_id"`
	MessageHash string `json:"message_hash"`
}

func NewHandler(validator ValidatorNode, network P2PNetwork) *Handler {
	return &Handler{
		validator: validator,
		network:   network,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:              h.validator.GetStatus(),
		Timestamp:           time.Now(),
		ValidatorAddress:    h.validator.GetAddress(),
		IsRegistered:        h.validator.IsRegistered(),
		Stake:               h.validator.GetStake(),
		PeerCount:           h.network.GetPeerCount(),
		PendingValidations:  h.validator.GetPendingValidationCount(),
		NetworkRunning:      h.network.IsRunning(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	response := StatusResponse{
		ValidatorAddress:   h.validator.GetAddress(),
		Status:             h.validator.GetStatus(),
		IsRegistered:       h.validator.IsRegistered(),
		Stake:              h.validator.GetStake(),
		PeerCount:          h.network.GetPeerCount(),
		PendingValidations: h.validator.GetPendingValidationCount(),
		NetworkRunning:     h.network.IsRunning(),
		Peers:              h.network.GetPeers(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) RequestValidation(w http.ResponseWriter, r *http.Request) {
	var payload ValidationRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	validationReq := &ValidationRequest{
		ID:           payload.PaymentID, // Use payment ID as validation ID for simplicity
		PaymentID:    payload.PaymentID,
		MessageHash:  payload.MessageHash,
		RequiredSigs: payload.RequiredSigs,
		Deadline:     time.Now().Add(5 * time.Minute),
		IsHighValue:  payload.IsHighValue,
	}

	if err := h.validator.ProcessValidationRequest(validationReq); err != nil {
		http.Error(w, fmt.Sprintf("Failed to process validation request: %v", err), http.StatusInternalServerError)
		return
	}

	p2pMsg := &p2p.ValidationMessage{
		Type:        "validation_request",
		RequestID:   validationReq.ID,
		PaymentID:   validationReq.PaymentID,
		MessageHash: validationReq.MessageHash,
		Timestamp:   time.Now(),
	}

	if err := h.network.BroadcastValidationRequest(p2pMsg); err != nil {
		log.Printf("Failed to broadcast validation request: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"request_id": validationReq.ID,
		"status":     "requested",
		"deadline":   validationReq.Deadline,
	})
}

func (h *Handler) SignMessage(w http.ResponseWriter, r *http.Request) {
	var payload SignMessagePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	req, exists := h.validator.GetValidationStatus(payload.RequestID)
	if !exists {
		http.Error(w, "Validation request not found", http.StatusNotFound)
		return
	}

	signatures := h.validator.GetSignatures(payload.RequestID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"request_id":        req.ID,
		"payment_id":        req.PaymentID,
		"signatures_count":  len(signatures),
		"required_signatures": req.RequiredSigs,
		"signatures":        signatures,
		"deadline":          req.Deadline,
	})
}

func (h *Handler) GetPeers(w http.ResponseWriter, r *http.Request) {
	peers := h.network.GetPeers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"peer_count": len(peers),
		"peers":      peers,
	})
}

func (h *Handler) RegisterValidator(w http.ResponseWriter, r *http.Request) {
	if h.validator.IsRegistered() {
		http.Error(w, "Validator already registered", http.StatusConflict)
		return
	}

	stakeStr := r.URL.Query().Get("stake")
	if stakeStr == "" {
		http.Error(w, "Stake amount required", http.StatusBadRequest)
		return
	}

	_, err := strconv.ParseFloat(stakeStr, 64)
	if err != nil {
		http.Error(w, "Invalid stake amount", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "registration_requested",
		"address":  h.validator.GetAddress(),
		"stake":    stakeStr,
		"message":  "Registration transaction should be submitted to the RelayValidator contract",
	})
}