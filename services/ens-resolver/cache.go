package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

type CacheStats struct {
	ForwardEntries    int   `json:"forward_entries"`
	ReverseEntries    int   `json:"reverse_entries"`
	SubnameEntries    int   `json:"subname_entries"`
	TotalEntries      int   `json:"total_entries"`
	CacheHits         int64 `json:"cache_hits"`
	CacheMisses       int64 `json:"cache_misses"`
	HitRate           float64 `json:"hit_rate"`
	LastEviction      int64 `json:"last_eviction"`
	EvictedEntries    int64 `json:"evicted_entries"`
}

var (
	cacheHits      int64
	cacheMisses    int64
	lastEviction   int64
	evictedEntries int64
)

func initCache() {
	log.Println("Initializing ENS cache...")
	lastEviction = time.Now().Unix()
}

func handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	cacheMutex.RLock()
	forwardCount := len(ensCache)
	reverseCount := len(reverseCache)
	subnameCount := len(subnameRegistry)
	cacheMutex.RUnlock()
	
	totalRequests := cacheHits + cacheMisses
	var hitRate float64
	if totalRequests > 0 {
		hitRate = float64(cacheHits) / float64(totalRequests) * 100
	}
	
	stats := CacheStats{
		ForwardEntries:  forwardCount,
		ReverseEntries:  reverseCount,
		SubnameEntries:  subnameCount,
		TotalEntries:    forwardCount + reverseCount + subnameCount,
		CacheHits:       cacheHits,
		CacheMisses:     cacheMisses,
		HitRate:         hitRate,
		LastEviction:    lastEviction,
		EvictedEntries:  evictedEntries,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleClearCache(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	cacheMutex.Lock()
	
	forwardCount := len(ensCache)
	reverseCount := len(reverseCache)
	subnameCount := len(subnameRegistry)
	
	ensCache = make(map[string]ENSRecord)
	reverseCache = make(map[string]ReverseRecord)
	subnameRegistry = make(map[string][]string)
	
	cacheMutex.Unlock()
	
	// Reset stats
	cacheHits = 0
	cacheMisses = 0
	evictedEntries = 0
	
	totalCleared := forwardCount + reverseCount + subnameCount
	
	log.Printf("Cache cleared: %d total entries removed", totalCleared)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Cache cleared successfully",
		"cleared_entries": map[string]interface{}{
			"forward":  forwardCount,
			"reverse":  reverseCount,
			"subnames": subnameCount,
			"total":    totalCleared,
		},
		"timestamp": time.Now().Unix(),
	})
}

func handleClearCacheEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	path := strings.TrimPrefix(r.URL.Path, "/api/cache/entry/")
	key := path
	if key == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Cache key required"})
		return
	}
	
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	removed := 0
	
	// Try to remove from forward cache
	if _, exists := ensCache[key]; exists {
		delete(ensCache, key)
		removed++
	}
	
	// Try to remove from reverse cache
	if _, exists := reverseCache[key]; exists {
		delete(reverseCache, key)
		removed++
	}
	
	// Try to remove from subname registry
	if _, exists := subnameRegistry[key]; exists {
		delete(subnameRegistry, key)
		removed++
	}
	
	if removed > 0 {
		log.Printf("Cache entry removed: %s (%d entries)", key, removed)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Cache entry removed",
			"key":     key,
			"removed": removed,
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "Cache entry not found",
			"key":     key,
		})
	}
}

func evictExpiredEntries() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	now := time.Now().Unix()
	evicted := 0
	
	// Evict expired forward entries
	for name, record := range ensCache {
		if now-record.Timestamp > record.TTL {
			delete(ensCache, name)
			evicted++
		}
	}
	
	// Evict expired reverse entries
	for addr, record := range reverseCache {
		if now-record.Timestamp > record.TTL {
			delete(reverseCache, addr)
			evicted++
		}
	}
	
	if evicted > 0 {
		evictedEntries += int64(evicted)
		lastEviction = now
		log.Printf("Cache eviction: %d expired entries removed", evicted)
	}
}

func recordCacheHit() {
	cacheHits++
}

func recordCacheMiss() {
	cacheMisses++
}