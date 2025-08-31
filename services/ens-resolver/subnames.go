package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type SubnameRegistration struct {
	Subname     string            `json:"subname"`
	Domain      string            `json:"domain"`
	Owner       string            `json:"owner"`
	Address     string            `json:"address"`
	TextRecords map[string]string `json:"text_records,omitempty"`
	CreatedAt   int64             `json:"created_at"`
	ExpiresAt   int64             `json:"expires_at"`
	Active      bool              `json:"active"`
}

type BulkRegistrationRequest struct {
	Domain      string                    `json:"domain" binding:"required"`
	Owner       string                    `json:"owner" binding:"required"`
	Subnames    []string                  `json:"subnames" binding:"required"`
	DefaultTTL  int64                     `json:"default_ttl"`
	TextRecords map[string]string         `json:"text_records,omitempty"`
}

type BulkRegistrationResponse struct {
	Successful []string `json:"successful"`
	Failed     []string `json:"failed"`
	Total      int      `json:"total"`
	Errors     []string `json:"errors,omitempty"`
}

var (
	subnameRegistry = make(map[string][]string) // domain -> list of subnames
	subnameRecords  = make(map[string]SubnameRegistration) // full_subname -> registration
)

func initSubnameRegistry() {
	log.Println("Initializing subname registry...")
	
	// Pre-populate with some mock subnames
	mockSubnames := []SubnameRegistration{
		{
			Subname: "pay.crosspay.eth",
			Domain:  "crosspay.eth",
			Owner:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			Address: "0x1111111111111111111111111111111111111111",
			TextRecords: map[string]string{
				"description": "Payment gateway for CrossPay",
				"url":         "https://pay.crosspay.xyz",
			},
			CreatedAt: time.Now().Unix() - 3600,
			ExpiresAt: time.Now().Unix() + 31536000, // 1 year
			Active:    true,
		},
		{
			Subname: "api.crosspay.eth",
			Domain:  "crosspay.eth", 
			Owner:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
			Address: "0x2222222222222222222222222222222222222222",
			TextRecords: map[string]string{
				"description": "CrossPay API endpoint",
				"url":         "https://api.crosspay.xyz",
			},
			CreatedAt: time.Now().Unix() - 1800,
			ExpiresAt: time.Now().Unix() + 31536000,
			Active:    true,
		},
	}
	
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	for _, reg := range mockSubnames {
		subnameRecords[reg.Subname] = reg
		
		if existing, ok := subnameRegistry[reg.Domain]; ok {
			subnameRegistry[reg.Domain] = append(existing, reg.Subname)
		} else {
			subnameRegistry[reg.Domain] = []string{reg.Subname}
		}
	}
	
	log.Printf("Subname registry initialized with %d records", len(subnameRecords))
}

func handleRegisterSubname(c *gin.Context) {
	var request struct {
		Subname     string            `json:"subname" binding:"required"`
		Domain      string            `json:"domain" binding:"required"`
		Owner       string            `json:"owner" binding:"required"`
		Address     string            `json:"address" binding:"required"`
		TTL         int64             `json:"ttl"`
		TextRecords map[string]string `json:"text_records"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Validate inputs
	if !strings.HasSuffix(request.Domain, ".eth") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .eth domains supported"})
		return
	}
	
	if strings.Contains(request.Subname, ".") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subname should not contain dots"})
		return
	}
	
	if !isValidAddress(request.Owner) || !isValidAddress(request.Address) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address format"})
		return
	}
	
	fullSubname := request.Subname + "." + request.Domain
	
	// Check if subname already exists
	cacheMutex.RLock()
	existing, exists := subnameRecords[fullSubname]
	cacheMutex.RUnlock()
	
	if exists && existing.Active {
		c.JSON(http.StatusConflict, gin.H{"error": "Subname already registered"})
		return
	}
	
	// Set default TTL if not provided
	ttl := request.TTL
	if ttl <= 0 {
		ttl = 31536000 // 1 year default
	}
	
	// Create registration
	registration := SubnameRegistration{
		Subname:     fullSubname,
		Domain:      request.Domain,
		Owner:       request.Owner,
		Address:     request.Address,
		TextRecords: request.TextRecords,
		CreatedAt:   time.Now().Unix(),
		ExpiresAt:   time.Now().Unix() + ttl,
		Active:      true,
	}
	
	// Store registration
	cacheMutex.Lock()
	subnameRecords[fullSubname] = registration
	
	if existing, ok := subnameRegistry[request.Domain]; ok {
		// Add to existing domain's subnames if not already present
		found := false
		for _, sub := range existing {
			if sub == fullSubname {
				found = true
				break
			}
		}
		if !found {
			subnameRegistry[request.Domain] = append(existing, fullSubname)
		}
	} else {
		subnameRegistry[request.Domain] = []string{fullSubname}
	}
	
	// Also add to ENS cache for resolution
	ensCache[fullSubname] = ENSRecord{
		Name:        fullSubname,
		Address:     request.Address,
		TextRecords: request.TextRecords,
		Timestamp:   time.Now().Unix(),
		TTL:         3600, // 1 hour cache TTL
	}
	cacheMutex.Unlock()
	
	log.Printf("Subname registered: %s -> %s", fullSubname, request.Address)
	
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Subname registered successfully",
		"subname":     fullSubname,
		"address":     request.Address,
		"owner":       request.Owner,
		"expires_at":  registration.ExpiresAt,
		"created_at":  registration.CreatedAt,
	})
}

func handleListSubnames(c *gin.Context) {
	domain := strings.ToLower(c.Param("domain"))
	
	if !strings.HasSuffix(domain, ".eth") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .eth domains supported"})
		return
	}
	
	cacheMutex.RLock()
	subnames, exists := subnameRegistry[domain]
	
	if !exists {
		cacheMutex.RUnlock()
		c.JSON(http.StatusOK, gin.H{
			"domain":   domain,
			"subnames": []string{},
			"count":    0,
		})
		return
	}
	
	// Get full registration details
	var registrations []SubnameRegistration
	for _, subname := range subnames {
		if reg, ok := subnameRecords[subname]; ok && reg.Active {
			registrations = append(registrations, reg)
		}
	}
	cacheMutex.RUnlock()
	
	c.JSON(http.StatusOK, gin.H{
		"domain":        domain,
		"registrations": registrations,
		"count":         len(registrations),
	})
}

func handleBulkRegister(c *gin.Context) {
	var request BulkRegistrationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	if len(request.Subnames) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No subnames provided"})
		return
	}
	
	if len(request.Subnames) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many subnames (max 100)"})
		return
	}
	
	// Validate domain and owner
	if !strings.HasSuffix(request.Domain, ".eth") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .eth domains supported"})
		return
	}
	
	if !isValidAddress(request.Owner) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid owner address"})
		return
	}
	
	// Set default TTL
	ttl := request.DefaultTTL
	if ttl <= 0 {
		ttl = 31536000 // 1 year
	}
	
	var successful []string
	var failed []string
	var errors []string
	
	now := time.Now().Unix()
	
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	for i, subname := range request.Subnames {
		if strings.Contains(subname, ".") {
			failed = append(failed, subname)
			errors = append(errors, fmt.Sprintf("Subname '%s' contains invalid characters", subname))
			continue
		}
		
		fullSubname := subname + "." + request.Domain
		
		// Check if already exists
		if existing, exists := subnameRecords[fullSubname]; exists && existing.Active {
			failed = append(failed, subname)
			errors = append(errors, fmt.Sprintf("Subname '%s' already registered", subname))
			continue
		}
		
		// Generate address for this subname (mock - would be provided or generated)
		address := fmt.Sprintf("0x%040d", 1000+i) // Mock address generation
		
		// Create registration
		registration := SubnameRegistration{
			Subname:     fullSubname,
			Domain:      request.Domain,
			Owner:       request.Owner,
			Address:     address,
			TextRecords: request.TextRecords,
			CreatedAt:   now,
			ExpiresAt:   now + ttl,
			Active:      true,
		}
		
		// Store registration
		subnameRecords[fullSubname] = registration
		
		// Update domain registry
		if existing, ok := subnameRegistry[request.Domain]; ok {
			subnameRegistry[request.Domain] = append(existing, fullSubname)
		} else {
			subnameRegistry[request.Domain] = []string{fullSubname}
		}
		
		// Add to ENS cache
		ensCache[fullSubname] = ENSRecord{
			Name:        fullSubname,
			Address:     address,
			TextRecords: request.TextRecords,
			Timestamp:   now,
			TTL:         3600,
		}
		
		successful = append(successful, subname)
	}
	
	log.Printf("Bulk registration for %s: %d successful, %d failed", 
		request.Domain, len(successful), len(failed))
	
	response := BulkRegistrationResponse{
		Successful: successful,
		Failed:     failed,
		Total:      len(request.Subnames),
		Errors:     errors,
	}
	
	c.JSON(http.StatusOK, response)
}

func handleRevokeSubname(c *gin.Context) {
	subname := strings.ToLower(c.Param("subname"))
	
	if !strings.Contains(subname, ".") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subname format"})
		return
	}
	
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	registration, exists := subnameRecords[subname]
	if !exists || !registration.Active {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subname not found or already inactive"})
		return
	}
	
	// Deactivate registration
	registration.Active = false
	subnameRecords[subname] = registration
	
	// Remove from ENS cache
	delete(ensCache, subname)
	
	// Remove from reverse cache if exists
	delete(reverseCache, registration.Address)
	
	log.Printf("Subname revoked: %s", subname)
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Subname revoked successfully",
		"subname":    subname,
		"revoked_at": time.Now().Unix(),
	})
}