package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RandomRequest struct {
	ID        string    `json:"id"`
	Requester string    `json:"requester"`
	Timestamp int64     `json:"timestamp"`
	Status    string    `json:"status"` // "pending", "fulfilled"
	Seed      string    `json:"seed,omitempty"`
	FulfilledAt int64   `json:"fulfilled_at,omitempty"`
}

var (
	randomRequests = make(map[string]*RandomRequest)
	randomMutex    = sync.RWMutex{}
	requestCounter = 0
)

func initializeRNG() {
	log.Println("Initializing RNG service...")
	// Mock initialization - would connect to Flare RNG service
	log.Println("RNG service initialized")
}

func handleRequestRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var request struct {
		Requester string `json:"requester,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// Ignore decode error, use anonymous as default
	}
	
	if request.Requester == "" {
		request.Requester = "anonymous"
	}
	
	randomMutex.Lock()
	requestCounter++
	requestID := fmt.Sprintf("rng_%d_%d", time.Now().Unix(), requestCounter)
	
	randomReq := &RandomRequest{
		ID:        requestID,
		Requester: request.Requester,
		Timestamp: time.Now().Unix(),
		Status:    "pending",
	}
	
	randomRequests[requestID] = randomReq
	randomMutex.Unlock()
	
	log.Printf("Random number requested: %s by %s", requestID, request.Requester)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"request_id": requestID,
		"status":     "pending",
		"timestamp":  randomReq.Timestamp,
		"estimated_fulfillment": time.Now().Unix() + 60, // 1 minute delay
	})
}

func handleRandomStatus(w http.ResponseWriter, r *http.Request) {
	// Extract requestId from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/random/status/")
	requestID := strings.TrimSuffix(path, "/")
	
	randomMutex.RLock()
	request, exists := randomRequests[requestID]
	randomMutex.RUnlock()
	
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Request not found"})
		return
	}
	
	response := map[string]interface{}{
		"request_id": request.ID,
		"status":     request.Status,
		"timestamp":  request.Timestamp,
		"requester":  request.Requester,
	}
	
	if request.Status == "fulfilled" {
		response["seed"] = request.Seed
		response["fulfilled_at"] = request.FulfilledAt
	} else {
		// Estimate fulfillment time
		elapsed := time.Now().Unix() - request.Timestamp
		remaining := 60 - elapsed // 1 minute fulfillment delay
		if remaining < 0 {
			remaining = 0
		}
		response["estimated_seconds_remaining"] = remaining
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleFulfillRandom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var request struct {
		RequestID string `json:"request_id"`
		Seed      string `json:"seed,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid request format"})
		return
	}

	if request.RequestID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Request ID is required"})
		return
	}
	
	randomMutex.Lock()
	defer randomMutex.Unlock()
	
	randomReq, exists := randomRequests[request.RequestID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Request not found"})
		return
	}
	
	if randomReq.Status == "fulfilled" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Request already fulfilled"})
		return
	}
	
	// Check if enough time has passed (minimum 1 minute delay for security)
	if time.Now().Unix()-randomReq.Timestamp < 60 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooEarly)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Fulfillment too early",
			"minimum_wait_seconds": 60,
		})
		return
	}
	
	// Generate or use provided seed
	var seed string
	if request.Seed != "" {
		seed = request.Seed
	} else {
		// Generate cryptographically secure random seed
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": "Failed to generate random seed"})
			return
		}
		seed = hex.EncodeToString(randomBytes)
	}
	
	randomReq.Status = "fulfilled"
	randomReq.Seed = seed
	randomReq.FulfilledAt = time.Now().Unix()
	
	log.Printf("Random number fulfilled: %s with seed %s", request.RequestID, seed[:16]+"...")
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"request_id":   request.RequestID,
		"status":       "fulfilled",
		"seed":         seed,
		"fulfilled_at": randomReq.FulfilledAt,
	})
}

func fulfillPendingRandomRequests() {
	randomMutex.Lock()
	defer randomMutex.Unlock()
	
	now := time.Now().Unix()
	fulfilled := 0
	
	for requestID, request := range randomRequests {
		if request.Status == "pending" && now-request.Timestamp >= 60 {
			// Auto-fulfill after 1 minute
			randomBytes := make([]byte, 32)
			if _, err := rand.Read(randomBytes); err != nil {
				log.Printf("Failed to generate random seed for %s: %v", requestID, err)
				continue
			}
			
			seed := hex.EncodeToString(randomBytes)
			request.Status = "fulfilled"
			request.Seed = seed
			request.FulfilledAt = now
			
			fulfilled++
		}
	}
	
	if fulfilled > 0 {
		log.Printf("Auto-fulfilled %d pending random requests", fulfilled)
	}
}

// Helper function for grant selection and fair randomization
func selectRandomWinners(participants []string, numWinners int, seed string) ([]string, error) {
	if len(participants) == 0 {
		return nil, fmt.Errorf("no participants")
	}
	
	if numWinners >= len(participants) {
		return participants, nil
	}
	
	// Use seed to initialize deterministic random selection
	// This ensures the same seed always produces the same results
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return nil, fmt.Errorf("invalid seed format")
	}
	
	// Simple deterministic selection based on seed
	// In production, would use more sophisticated algorithm
	winners := make([]string, 0, numWinners)
	used := make(map[int]bool)
	
	for i := 0; i < numWinners && len(winners) < len(participants); i++ {
		// Generate index based on seed and iteration
		idx := int(seedBytes[i%len(seedBytes)]) % len(participants)
		
		// Ensure uniqueness
		attempts := 0
		for used[idx] && attempts < len(participants) {
			idx = (idx + 1) % len(participants)
			attempts++
		}
		
		if !used[idx] {
			winners = append(winners, participants[idx])
			used[idx] = true
		}
	}
	
	return winners, nil
}

// API endpoint for grant selection
func handleSelectWinners(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Method not allowed"})
		return
	}

	var request struct {
		Participants []string `json:"participants"`
		NumWinners   int      `json:"num_winners"`
		Seed         string   `json:"seed"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid request format"})
		return
	}

	if len(request.Participants) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Participants are required"})
		return
	}

	if request.NumWinners <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Number of winners must be greater than 0"})
		return
	}

	if request.Seed == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Seed is required"})
		return
	}
	
	winners, err := selectRandomWinners(request.Participants, request.NumWinners, request.Seed)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"winners":         winners,
		"total_participants": len(request.Participants),
		"num_winners":     len(winners),
		"seed_used":       request.Seed,
		"timestamp":       time.Now().Unix(),
	})
}