package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ENSRecord struct {
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Avatar    string            `json:"avatar,omitempty"`
	TextRecords map[string]string `json:"text_records,omitempty"`
	Timestamp int64             `json:"timestamp"`
	TTL       int64             `json:"ttl"`
}

type ReverseRecord struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	TTL       int64  `json:"ttl"`
}

type BatchResolveRequest struct {
	Names []string `json:"names" binding:"required"`
}

type BatchResolveResponse struct {
	Results []ENSRecord `json:"results"`
	Errors  []string    `json:"errors,omitempty"`
}

var (
	ensCache = make(map[string]ENSRecord)
	reverseCache = make(map[string]ReverseRecord)
	cacheMutex = sync.RWMutex{}
	
	// Mock ENS data for demonstration
	mockENSData = map[string]ENSRecord{
		"alice.eth": {
			Name:    "alice.eth",
			Address: "0x1234567890123456789012345678901234567890",
			Avatar:  "https://metadata.ens.domains/mainnet/avatar/alice.eth",
			TextRecords: map[string]string{
				"email":   "alice@example.com",
				"url":     "https://alice.example.com",
				"twitter": "@alice",
			},
			Timestamp: time.Now().Unix(),
			TTL:       3600,
		},
		"bob.eth": {
			Name:    "bob.eth", 
			Address: "0x0987654321098765432109876543210987654321",
			Avatar:  "https://metadata.ens.domains/mainnet/avatar/bob.eth",
			TextRecords: map[string]string{
				"email": "bob@example.com",
				"url":   "https://bob.example.com",
			},
			Timestamp: time.Now().Unix(),
			TTL:       3600,
		},
		"crosspay.eth": {
			Name:    "crosspay.eth",
			Address: "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			Avatar:  "",
			TextRecords: map[string]string{
				"description": "Cross-chain payment protocol",
				"url":         "https://crosspay.xyz",
			},
			Timestamp: time.Now().Unix(),
			TTL:       7200,
		},
	}
	
	mockReverseData = map[string]ReverseRecord{
		"0x1234567890123456789012345678901234567890": {
			Address:   "0x1234567890123456789012345678901234567890",
			Name:      "alice.eth",
			Timestamp: time.Now().Unix(),
			TTL:       3600,
		},
		"0x0987654321098765432109876543210987654321": {
			Address:   "0x0987654321098765432109876543210987654321",
			Name:      "bob.eth",
			Timestamp: time.Now().Unix(),
			TTL:       3600,
		},
	}
)

func initENSClient() {
	log.Println("Initializing ENS client...")
	
	// Pre-populate cache with mock data
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	for name, record := range mockENSData {
		ensCache[strings.ToLower(name)] = record
	}
	
	for addr, record := range mockReverseData {
		reverseCache[strings.ToLower(addr)] = record
	}
	
	log.Printf("ENS client initialized with %d forward and %d reverse records", 
		len(ensCache), len(reverseCache))
}

func handleResolveName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := strings.TrimPrefix(r.URL.Path, "/api/ens/resolve/")
	name := strings.ToLower(path)
	
	if !strings.HasSuffix(name, ".eth") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Only .eth domains supported"})
		return
	}
	
	// Check cache first
	cacheMutex.RLock()
	cached, exists := ensCache[name]
	cacheMutex.RUnlock()
	
	if exists && time.Now().Unix()-cached.Timestamp < cached.TTL {
		log.Printf("Cache hit for %s", name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}
	
	// Mock resolution (would query actual ENS)
	record, err := resolveENSName(name)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Name not found: %s", name)})
		return
	}
	
	// Update cache
	cacheMutex.Lock()
	ensCache[name] = record
	cacheMutex.Unlock()
	
	log.Printf("Resolved %s to %s", name, record.Address)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func handleReverseResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := strings.TrimPrefix(r.URL.Path, "/api/ens/reverse/")
	address := strings.ToLower(path)
	
	if !isValidAddress(address) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid address format"})
		return
	}
	
	// Check cache first
	cacheMutex.RLock()
	cached, exists := reverseCache[address]
	cacheMutex.RUnlock()
	
	if exists && time.Now().Unix()-cached.Timestamp < cached.TTL {
		log.Printf("Reverse cache hit for %s", address)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}
	
	// Mock reverse resolution
	record, err := reverseResolveAddress(address)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("No ENS name found for address: %s", address)})
		return
	}
	
	// Update cache
	cacheMutex.Lock()
	reverseCache[address] = record
	cacheMutex.Unlock()
	
	log.Printf("Reverse resolved %s to %s", address, record.Name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

func handleBatchResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var request BatchResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request format"})
		return
	}
	
	if len(request.Names) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "No names provided"})
		return
	}
	
	if len(request.Names) > 50 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Too many names (max 50)"})
		return
	}
	
	var results []ENSRecord
	var errors []string
	
	for _, name := range request.Names {
		normalizedName := strings.ToLower(name)
		
		if !strings.HasSuffix(normalizedName, ".eth") {
			errors = append(errors, fmt.Sprintf("Invalid name format: %s", name))
			continue
		}
		
		// Check cache
		cacheMutex.RLock()
		cached, exists := ensCache[normalizedName]
		cacheMutex.RUnlock()
		
		if exists && time.Now().Unix()-cached.Timestamp < cached.TTL {
			results = append(results, cached)
			continue
		}
		
		// Resolve name
		record, err := resolveENSName(normalizedName)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to resolve %s: %v", name, err))
			continue
		}
		
		// Update cache
		cacheMutex.Lock()
		ensCache[normalizedName] = record
		cacheMutex.Unlock()
		
		results = append(results, record)
	}
	
	response := BatchResolveResponse{
		Results: results,
		Errors:  errors,
	}
	
	log.Printf("Batch resolved %d names, %d errors", len(results), len(errors))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := strings.TrimPrefix(r.URL.Path, "/api/ens/avatar/")
	name := strings.ToLower(path)
	
	// Get ENS record
	cacheMutex.RLock()
	record, exists := ensCache[name]
	cacheMutex.RUnlock()
	
	if !exists {
		// Try to resolve first
		resolved, err := resolveENSName(name)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Name not found"})
			return
		}
		record = resolved
	}
	
	if record.Avatar == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No avatar set for this name"})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"name":   record.Name,
		"avatar": record.Avatar,
	})
}

func handleGetTextRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := strings.TrimPrefix(r.URL.Path, "/api/ens/text/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid path format"})
		return
	}
	name := strings.ToLower(parts[0])
	key := parts[1]
	
	// Get ENS record
	cacheMutex.RLock()
	record, exists := ensCache[name]
	cacheMutex.RUnlock()
	
	if !exists {
		// Try to resolve first
		resolved, err := resolveENSName(name)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Name not found"})
			return
		}
		record = resolved
	}
	
	if record.TextRecords == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No text records found"})
		return
	}
	
	value, exists := record.TextRecords[key]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Text record '%s' not found", key)})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":  record.Name,
		"key":   key,
		"value": value,
	})
}

func handleSearchNames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	query := strings.ToLower(r.URL.Query().Get("q"))
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Query parameter 'q' required"})
		return
	}
	
	var limit int = 20
	if l, err := parseLimit(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	
	var results []ENSRecord
	
	cacheMutex.RLock()
	for name, record := range ensCache {
		if len(results) >= limit {
			break
		}
		
		if strings.Contains(name, query) {
			results = append(results, record)
		}
	}
	cacheMutex.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
		"limit":   limit,
	})
}

func resolveENSName(name string) (ENSRecord, error) {
	// Mock resolution - would query actual ENS registry
	log.Printf("Resolving ENS name: %s", name)
	
	// Simulate network delay
	time.Sleep(50 * time.Millisecond)
	
	// Check mock data
	if record, exists := mockENSData[name]; exists {
		record.Timestamp = time.Now().Unix()
		return record, nil
	}
	
	return ENSRecord{}, fmt.Errorf("name not found: %s", name)
}

func reverseResolveAddress(address string) (ReverseRecord, error) {
	// Mock reverse resolution
	log.Printf("Reverse resolving address: %s", address)
	
	// Simulate network delay
	time.Sleep(50 * time.Millisecond)
	
	// Check mock data
	if record, exists := mockReverseData[address]; exists {
		record.Timestamp = time.Now().Unix()
		return record, nil
	}
	
	return ReverseRecord{}, fmt.Errorf("no ENS name for address: %s", address)
}

func isValidAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// Additional validation could be added here
	return true
}

func parseLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return 20, nil
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 20, fmt.Errorf("invalid limit: %v", err)
	}
	
	if limit <= 0 || limit > 100 {
		return 20, fmt.Errorf("limit must be between 1 and 100")
	}
	
	return limit, nil
}