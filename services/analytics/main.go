package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type AnalyticsServer struct {
	influxClient  influxdb2.Client
	writeAPI      api.WriteAPI
	queryAPI      api.QueryAPI
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]bool
	clientsMutex  sync.RWMutex
	paymentStream chan PaymentMetric
}

type PaymentMetric struct {
	PaymentID     uint64    `json:"payment_id"`
	ChainID       uint64    `json:"chain_id"`
	Sender        string    `json:"sender"`
	Recipient     string    `json:"recipient"`
	Token         string    `json:"token"`
	Amount        string    `json:"amount"`
	Fee           string    `json:"fee"`
	Status        string    `json:"status"`
	IsPrivate     bool      `json:"is_private"`
	RequiredSigs  uint32    `json:"required_sigs,omitempty"`
	ReceivedSigs  uint32    `json:"received_sigs,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	ProcessingTime int64     `json:"processing_time_ms,omitempty"`
}

type ValidatorMetric struct {
	ValidatorAddr string    `json:"validator_address"`
	ChainID       uint64    `json:"chain_id"`
	Stake         string    `json:"stake"`
	Status        string    `json:"status"`
	ResponseTime  int64     `json:"response_time_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

type VaultMetric struct {
	VaultAddress   string    `json:"vault_address"`
	ChainID        uint64    `json:"chain_id"`
	TrancheType    string    `json:"tranche_type"`
	TotalAssets    string    `json:"total_assets"`
	UtilizationPct float64   `json:"utilization_pct"`
	APY            float64   `json:"apy"`
	RiskScore      float64   `json:"risk_score"`
	SlashingEvents uint64    `json:"slashing_events"`
	Timestamp      time.Time `json:"timestamp"`
}

type AnalyticsQuery struct {
	MetricType string            `json:"metric_type"` // "payments", "validators", "vaults"
	TimeRange  string            `json:"time_range"`  // "1h", "24h", "7d", "30d"
	ChainID    *uint64           `json:"chain_id,omitempty"`
	Filters    map[string]string `json:"filters,omitempty"`
}

type AnalyticsResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   string      `json:"error,omitempty"`
}

func NewAnalyticsServer() *AnalyticsServer {
	influxURL := getEnv("INFLUXDB_URL", "http://localhost:8086")
	token := getEnv("INFLUXDB_TOKEN", "your-token-here")
	org := getEnv("INFLUXDB_ORG", "crosspay")
	bucket := getEnv("INFLUXDB_BUCKET", "analytics")

	client := influxdb2.NewClient(influxURL, token)
	writeAPI := client.WriteAPI(org, bucket)
	queryAPI := client.QueryAPI(org)

	return &AnalyticsServer{
		influxClient:  client,
		writeAPI:      writeAPI,
		queryAPI:      queryAPI,
		upgrader:      websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		clients:       make(map[*websocket.Conn]bool),
		paymentStream: make(chan PaymentMetric, 1000),
	}
}

func (s *AnalyticsServer) Start() {
	// Start background workers
	go s.processMetrics()
	go s.handleWebSocketBroadcasts()

	router := mux.NewRouter()

	// REST API endpoints
	router.HandleFunc("/api/metrics/payment", s.handlePaymentMetric).Methods("POST")
	router.HandleFunc("/api/metrics/validator", s.handleValidatorMetric).Methods("POST")
	router.HandleFunc("/api/metrics/vault", s.handleVaultMetric).Methods("POST")
	router.HandleFunc("/api/query", s.handleQuery).Methods("POST")
	router.HandleFunc("/api/dashboard", s.handleDashboard).Methods("GET")
	router.HandleFunc("/api/realtime/{metric_type}", s.handleRealtimeQuery).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws", s.handleWebSocket)

	// CORS middleware
	router.Use(corsMiddleware)

	port := getEnv("PORT", "8084")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Printf("Analytics server starting on port %s", port)

	// Graceful shutdown
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down analytics server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	s.influxClient.Close()
	log.Println("Analytics server stopped")
}

func (s *AnalyticsServer) handlePaymentMetric(w http.ResponseWriter, r *http.Request) {
	var metric PaymentMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Calculate processing time if completed
	if metric.CompletedAt != nil {
		metric.ProcessingTime = metric.CompletedAt.Sub(metric.Timestamp).Milliseconds()
	}

	// Send to processing channel
	select {
	case s.paymentStream <- metric:
	default:
		log.Printf("Payment stream channel full, dropping metric for payment %d", metric.PaymentID)
	}

	// Write to InfluxDB
	point := influxdb2.NewPointWithMeasurement("payments").
		AddTag("chain_id", fmt.Sprintf("%d", metric.ChainID)).
		AddTag("status", metric.Status).
		AddTag("token", metric.Token).
		AddTag("is_private", fmt.Sprintf("%t", metric.IsPrivate)).
		AddField("payment_id", metric.PaymentID).
		AddField("amount", metric.Amount).
		AddField("fee", metric.Fee).
		AddField("processing_time_ms", metric.ProcessingTime).
		SetTime(metric.Timestamp)

	if metric.RequiredSigs > 0 {
		point.AddField("required_sigs", metric.RequiredSigs).
			AddField("received_sigs", metric.ReceivedSigs)
	}

	s.writeAPI.WritePoint(point)

	// Broadcast to WebSocket clients
	s.broadcastToClients(map[string]interface{}{
		"type": "payment",
		"data": metric,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AnalyticsResponse{Success: true})
}

func (s *AnalyticsServer) handleValidatorMetric(w http.ResponseWriter, r *http.Request) {
	var metric ValidatorMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Write to InfluxDB
	point := influxdb2.NewPointWithMeasurement("validators").
		AddTag("chain_id", fmt.Sprintf("%d", metric.ChainID)).
		AddTag("validator_address", metric.ValidatorAddr).
		AddTag("status", metric.Status).
		AddField("stake", metric.Stake).
		AddField("response_time_ms", metric.ResponseTime).
		SetTime(metric.Timestamp)

	s.writeAPI.WritePoint(point)

	// Broadcast to WebSocket clients
	s.broadcastToClients(map[string]interface{}{
		"type": "validator",
		"data": metric,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AnalyticsResponse{Success: true})
}

func (s *AnalyticsServer) handleVaultMetric(w http.ResponseWriter, r *http.Request) {
	var metric VaultMetric
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Write to InfluxDB
	point := influxdb2.NewPointWithMeasurement("vaults").
		AddTag("chain_id", fmt.Sprintf("%d", metric.ChainID)).
		AddTag("vault_address", metric.VaultAddress).
		AddTag("tranche_type", metric.TrancheType).
		AddField("total_assets", metric.TotalAssets).
		AddField("utilization_pct", metric.UtilizationPct).
		AddField("apy", metric.APY).
		AddField("risk_score", metric.RiskScore).
		AddField("slashing_events", metric.SlashingEvents).
		SetTime(metric.Timestamp)

	s.writeAPI.WritePoint(point)

	// Broadcast to WebSocket clients
	s.broadcastToClients(map[string]interface{}{
		"type": "vault",
		"data": metric,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AnalyticsResponse{Success: true})
}

func (s *AnalyticsServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	var query AnalyticsQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	timeFilter := parseTimeRange(query.TimeRange)
	var fluxQuery string

	switch query.MetricType {
	case "payments":
		fluxQuery = fmt.Sprintf(`
			from(bucket: "analytics")
			|> range(start: %s)
			|> filter(fn: (r) => r["_measurement"] == "payments")
		`, timeFilter)
		
		if query.ChainID != nil {
			fluxQuery += fmt.Sprintf(`|> filter(fn: (r) => r["chain_id"] == "%d")`, *query.ChainID)
		}

	case "validators":
		fluxQuery = fmt.Sprintf(`
			from(bucket: "analytics")
			|> range(start: %s)
			|> filter(fn: (r) => r["_measurement"] == "validators")
		`, timeFilter)
		
		if query.ChainID != nil {
			fluxQuery += fmt.Sprintf(`|> filter(fn: (r) => r["chain_id"] == "%d")`, *query.ChainID)
		}

	case "vaults":
		fluxQuery = fmt.Sprintf(`
			from(bucket: "analytics")
			|> range(start: %s)
			|> filter(fn: (r) => r["_measurement"] == "vaults")
		`, timeFilter)
		
		if query.ChainID != nil {
			fluxQuery += fmt.Sprintf(`|> filter(fn: (r) => r["chain_id"] == "%d")`, *query.ChainID)
		}

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	// Execute query
	result, err := s.queryAPI.Query(context.Background(), fluxQuery)
	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	var records []map[string]interface{}
	for result.Next() {
		record := make(map[string]interface{})
		for key, value := range result.Record().Values() {
			record[key] = value
		}
		records = append(records, record)
	}

	if result.Err() != nil {
		log.Printf("Query result error: %v", result.Err())
		http.Error(w, "Query processing failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AnalyticsResponse{
		Success: true,
		Data:    records,
	})
}

func (s *AnalyticsServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get comprehensive dashboard data
	dashboardData := make(map[string]interface{})

	// Payment volume (last 24h)
	paymentQuery := `
		from(bucket: "analytics")
		|> range(start: -24h)
		|> filter(fn: (r) => r["_measurement"] == "payments")
		|> group(columns: ["status"])
		|> count()
	`
	
	paymentResult, err := s.queryAPI.Query(context.Background(), paymentQuery)
	if err == nil {
		paymentStats := make(map[string]int64)
		for paymentResult.Next() {
			status := paymentResult.Record().ValueByKey("status").(string)
			count := paymentResult.Record().Value().(int64)
			paymentStats[status] = count
		}
		dashboardData["payment_stats"] = paymentStats
	}

	// Validator health
	validatorQuery := `
		from(bucket: "analytics")
		|> range(start: -1h)
		|> filter(fn: (r) => r["_measurement"] == "validators")
		|> group(columns: ["status"])
		|> count()
	`
	
	validatorResult, err := s.queryAPI.Query(context.Background(), validatorQuery)
	if err == nil {
		validatorStats := make(map[string]int64)
		for validatorResult.Next() {
			status := validatorResult.Record().ValueByKey("status").(string)
			count := validatorResult.Record().Value().(int64)
			validatorStats[status] = count
		}
		dashboardData["validator_stats"] = validatorStats
	}

	// Vault metrics
	vaultQuery := `
		from(bucket: "analytics")
		|> range(start: -1h)
		|> filter(fn: (r) => r["_measurement"] == "vaults")
		|> last()
		|> group(columns: ["tranche_type"])
		|> mean(column: "_value")
	`
	
	vaultResult, err := s.queryAPI.Query(context.Background(), vaultQuery)
	if err == nil {
		vaultStats := make(map[string]float64)
		for vaultResult.Next() {
			tranche := vaultResult.Record().ValueByKey("tranche_type").(string)
			avgValue := vaultResult.Record().Value().(float64)
			vaultStats[tranche] = avgValue
		}
		dashboardData["vault_stats"] = vaultStats
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AnalyticsResponse{
		Success: true,
		Data:    dashboardData,
	})
}

func (s *AnalyticsServer) handleRealtimeQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	metricType := vars["metric_type"]

	// Get real-time data (last 5 minutes)
	timeFilter := "-5m"
	var fluxQuery string

	switch metricType {
	case "payments":
		fluxQuery = fmt.Sprintf(`
			from(bucket: "analytics")
			|> range(start: %s)
			|> filter(fn: (r) => r["_measurement"] == "payments")
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: 100)
		`, timeFilter)

	case "validators":
		fluxQuery = fmt.Sprintf(`
			from(bucket: "analytics")
			|> range(start: %s)
			|> filter(fn: (r) => r["_measurement"] == "validators")
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: 50)
		`, timeFilter)

	case "vaults":
		fluxQuery = fmt.Sprintf(`
			from(bucket: "analytics")
			|> range(start: %s)
			|> filter(fn: (r) => r["_measurement"] == "vaults")
			|> sort(columns: ["_time"], desc: true)
			|> limit(n: 20)
		`, timeFilter)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	result, err := s.queryAPI.Query(context.Background(), fluxQuery)
	if err != nil {
		log.Printf("Realtime query error: %v", err)
		http.Error(w, "Query failed", http.StatusInternalServerError)
		return
	}

	var records []map[string]interface{}
	for result.Next() {
		record := make(map[string]interface{})
		for key, value := range result.Record().Values() {
			record[key] = value
		}
		records = append(records, record)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AnalyticsResponse{
		Success: true,
		Data:    records,
	})
}

func (s *AnalyticsServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	s.clientsMutex.Lock()
	s.clients[conn] = true
	s.clientsMutex.Unlock()

	log.Printf("New WebSocket client connected. Total clients: %d", len(s.clients))

	// Handle client disconnection
	defer func() {
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
		log.Printf("WebSocket client disconnected. Remaining clients: %d", len(s.clients))
	}()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *AnalyticsServer) processMetrics() {
	for metric := range s.paymentStream {
		// Additional processing logic can be added here
		log.Printf("Processed payment metric: ID=%d, Chain=%d, Status=%s", 
			metric.PaymentID, metric.ChainID, metric.Status)
	}
}

func (s *AnalyticsServer) handleWebSocketBroadcasts() {
	// This goroutine handles broadcasting to WebSocket clients
	// In a real implementation, this would be triggered by the broadcastToClients method
	// For now, it just keeps the goroutine alive
	for {
		time.Sleep(time.Minute)
	}
}

func (s *AnalyticsServer) broadcastToClients(data map[string]interface{}) {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	message, _ := json.Marshal(data)
	
	for client := range s.clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

func parseTimeRange(timeRange string) string {
	switch strings.ToLower(timeRange) {
	case "1h":
		return "-1h"
	case "24h":
		return "-24h"
	case "7d":
		return "-7d"
	case "30d":
		return "-30d"
	default:
		return "-1h"
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	server := NewAnalyticsServer()
	server.Start()
}