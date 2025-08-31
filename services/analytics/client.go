package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// AnalyticsClient provides methods to send metrics to the analytics service
type AnalyticsClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAnalyticsClient creates a new analytics client
func NewAnalyticsClient(baseURL string) *AnalyticsClient {
	return &AnalyticsClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendPaymentMetric sends a payment metric to the analytics service
func (c *AnalyticsClient) SendPaymentMetric(metric PaymentMetric) error {
	return c.sendMetric("/api/metrics/payment", metric)
}

// SendValidatorMetric sends a validator metric to the analytics service
func (c *AnalyticsClient) SendValidatorMetric(metric ValidatorMetric) error {
	return c.sendMetric("/api/metrics/validator", metric)
}

// SendVaultMetric sends a vault metric to the analytics service
func (c *AnalyticsClient) SendVaultMetric(metric VaultMetric) error {
	return c.sendMetric("/api/metrics/vault", metric)
}

// QueryMetrics queries metrics from the analytics service
func (c *AnalyticsClient) QueryMetrics(query AnalyticsQuery) (*AnalyticsResponse, error) {
	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/query", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result AnalyticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetDashboard retrieves dashboard data
func (c *AnalyticsClient) GetDashboard() (*AnalyticsResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/dashboard")
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result AnalyticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetRealtimeMetrics retrieves real-time metrics for a specific type
func (c *AnalyticsClient) GetRealtimeMetrics(metricType string) (*AnalyticsResponse, error) {
	resp, err := c.HTTPClient.Get(fmt.Sprintf("%s/api/realtime/%s", c.BaseURL, metricType))
	if err != nil {
		return nil, fmt.Errorf("failed to get realtime metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var result AnalyticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// sendMetric is a helper method to send metrics to the analytics service
func (c *AnalyticsClient) sendMetric(endpoint string, metric interface{}) error {
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// Example usage and integration functions
func ExampleUsage() {
	// Create analytics client
	client := NewAnalyticsClient("http://localhost:8084")

	// Send payment metric
	paymentMetric := PaymentMetric{
		PaymentID:      12345,
		ChainID:        1,
		Sender:         "0x742d35Cc6634C0532925a3b8D4ba9f4e6ad1B6AF",
		Recipient:      "0x8ba1f109551bD432803012645Hac136c4c5688dC",
		Token:          "0x0000000000000000000000000000000000000000",
		Amount:         "1000000000000000000", // 1 ETH in wei
		Fee:            "1000000000000000",    // 0.001 ETH in wei
		Status:         "completed",
		IsPrivate:      false,
		RequiredSigs:   0,
		ReceivedSigs:   0,
		Timestamp:      time.Now(),
		CompletedAt:    timePtr(time.Now()),
		ProcessingTime: 15000, // 15 seconds in ms
	}

	if err := client.SendPaymentMetric(paymentMetric); err != nil {
		log.Printf("Failed to send payment metric: %v", err)
	}

	// Send validator metric
	validatorMetric := ValidatorMetric{
		ValidatorAddr: "0x1234567890123456789012345678901234567890",
		ChainID:       1,
		Stake:         "15000000000000000000", // 15 ETH
		Status:        "active",
		ResponseTime:  250, // 250ms
		Timestamp:     time.Now(),
	}

	if err := client.SendValidatorMetric(validatorMetric); err != nil {
		log.Printf("Failed to send validator metric: %v", err)
	}

	// Send vault metric
	vaultMetric := VaultMetric{
		VaultAddress:   "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		ChainID:        1,
		TrancheType:    "senior",
		TotalAssets:    "5000000000000000000000", // 5000 ETH
		UtilizationPct: 85.5,
		APY:            8.5,
		RiskScore:      0.2,
		SlashingEvents: 0,
		Timestamp:      time.Now(),
	}

	if err := client.SendVaultMetric(vaultMetric); err != nil {
		log.Printf("Failed to send vault metric: %v", err)
	}

	// Query payment metrics for the last 24 hours
	query := AnalyticsQuery{
		MetricType: "payments",
		TimeRange:  "24h",
		ChainID:    uint64Ptr(1),
		Filters: map[string]string{
			"status": "completed",
		},
	}

	result, err := client.QueryMetrics(query)
	if err != nil {
		log.Printf("Failed to query metrics: %v", err)
	} else {
		log.Printf("Query result: %+v", result.Data)
	}

	// Get dashboard data
	dashboard, err := client.GetDashboard()
	if err != nil {
		log.Printf("Failed to get dashboard: %v", err)
	} else {
		log.Printf("Dashboard data: %+v", dashboard.Data)
	}

	// Get real-time payment metrics
	realtimePayments, err := client.GetRealtimeMetrics("payments")
	if err != nil {
		log.Printf("Failed to get realtime payments: %v", err)
	} else {
		log.Printf("Realtime payments: %+v", realtimePayments.Data)
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

// IntegrationWithPaymentCore demonstrates how to integrate with PaymentCore
func IntegrationWithPaymentCore(client *AnalyticsClient, paymentID uint64, chainID uint64, sender, recipient, token string, amount, fee string, isPrivate bool) {
	// When a payment is created
	metric := PaymentMetric{
		PaymentID: paymentID,
		ChainID:   chainID,
		Sender:    sender,
		Recipient: recipient,
		Token:     token,
		Amount:    amount,
		Fee:       fee,
		Status:    "pending",
		IsPrivate: isPrivate,
		Timestamp: time.Now(),
	}

	if err := client.SendPaymentMetric(metric); err != nil {
		log.Printf("Failed to send payment created metric: %v", err)
	}
}

// IntegrationWithValidator demonstrates validator integration
func IntegrationWithValidator(client *AnalyticsClient, validatorAddr string, chainID uint64, stake string, responseTime int64) {
	metric := ValidatorMetric{
		ValidatorAddr: validatorAddr,
		ChainID:       chainID,
		Stake:         stake,
		Status:        "active",
		ResponseTime:  responseTime,
		Timestamp:     time.Now(),
	}

	if err := client.SendValidatorMetric(metric); err != nil {
		log.Printf("Failed to send validator metric: %v", err)
	}
}