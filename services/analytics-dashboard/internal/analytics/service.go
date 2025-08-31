package analytics

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/crosspay/analytics-dashboard/internal/metrics"
)

type Service struct {
	collector MetricsCollector
}

type MetricsCollector interface {
	GetValidatorMetrics() map[string]*metrics.ValidatorMetrics
	GetVaultMetrics() *metrics.VaultMetrics
	GetPaymentMetrics() *metrics.PaymentMetrics
	GetPrivacyMetrics() *metrics.PrivacyMetrics
	GetNetworkMetrics() *metrics.NetworkMetrics
	IsCollecting() bool
}

type DashboardResponse struct {
	Timestamp        time.Time                           `json:"timestamp"`
	ValidatorMetrics map[string]*metrics.ValidatorMetrics `json:"validator_metrics"`
	VaultMetrics     *metrics.VaultMetrics               `json:"vault_metrics"`
	PaymentMetrics   *metrics.PaymentMetrics             `json:"payment_metrics"`
	PrivacyMetrics   *metrics.PrivacyMetrics             `json:"privacy_metrics"`
	NetworkMetrics   *metrics.NetworkMetrics             `json:"network_metrics"`
	SystemStatus     string                              `json:"system_status"`
}

func NewService(collector MetricsCollector) *Service {
	return &Service{
		collector: collector,
	}
}

func (s *Service) GetMetrics(w http.ResponseWriter, r *http.Request) {
	response := DashboardResponse{
		Timestamp:        time.Now(),
		ValidatorMetrics: s.collector.GetValidatorMetrics(),
		VaultMetrics:     s.collector.GetVaultMetrics(),
		PaymentMetrics:   s.collector.GetPaymentMetrics(),
		PrivacyMetrics:   s.collector.GetPrivacyMetrics(),
		NetworkMetrics:   s.collector.GetNetworkMetrics(),
		SystemStatus:     s.getSystemStatus(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) GetValidatorMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.collector.GetValidatorMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": time.Now(),
		"validators": metrics,
		"summary": map[string]interface{}{
			"total_validators": len(metrics),
			"active_count":     s.countActiveValidators(metrics),
			"average_uptime":   s.calculateAverageUptime(metrics),
		},
	})
}

func (s *Service) GetVaultMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.collector.GetVaultMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": time.Now(),
		"vault":     metrics,
		"health": map[string]interface{}{
			"is_balanced":     s.isVaultBalanced(metrics),
			"risk_level":      s.calculateRiskLevel(metrics),
			"yield_trending":  "stable",
		},
	})
}

func (s *Service) GetPaymentMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.collector.GetPaymentMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": time.Now(),
		"payments":  metrics,
		"trends": map[string]interface{}{
			"hourly_volume":  "increasing",
			"privacy_adoption": (float64(metrics.PrivatePayments) / float64(metrics.TotalPayments)) * 100,
			"validation_performance": metrics.SuccessRate,
		},
	})
}

func (s *Service) GetPrivacyMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := s.collector.GetPrivacyMetrics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": time.Now(),
		"privacy":   metrics,
		"insights": map[string]interface{}{
			"disclosure_approval_rate": (float64(metrics.ApprovedDisclosures) / float64(metrics.DisclosureRequests)) * 100,
			"grant_participation":      metrics.SealedBidGrants,
			"privacy_trend":            "growing",
		},
	})
}

func (s *Service) getSystemStatus() string {
	if !s.collector.IsCollecting() {
		return "degraded"
	}

	networkMetrics := s.collector.GetNetworkMetrics()
	if networkMetrics.NetworkUptime < 95.0 {
		return "degraded"
	}

	validatorMetrics := s.collector.GetValidatorMetrics()
	if len(validatorMetrics) < 3 {
		return "warning"
	}

	return "healthy"
}

func (s *Service) countActiveValidators(validators map[string]*metrics.ValidatorMetrics) int {
	active := 0
	for _, v := range validators {
		if v.Status == "active" {
			active++
		}
	}
	return active
}

func (s *Service) calculateAverageUptime(validators map[string]*metrics.ValidatorMetrics) float64 {
	if len(validators) == 0 {
		return 0.0
	}

	totalUptime := 0.0
	for _, v := range validators {
		totalUptime += v.Uptime
	}
	
	return totalUptime / float64(len(validators))
}

func (s *Service) isVaultBalanced(vault *metrics.VaultMetrics) bool {
	return vault.UtilizationRates["junior"] >= 15.0 && 
		   vault.UtilizationRates["junior"] <= 25.0 &&
		   vault.UtilizationRates["mezzanine"] >= 25.0 && 
		   vault.UtilizationRates["mezzanine"] <= 35.0 &&
		   vault.UtilizationRates["senior"] >= 45.0 && 
		   vault.UtilizationRates["senior"] <= 55.0
}

func (s *Service) calculateRiskLevel(vault *metrics.VaultMetrics) string {
	if len(vault.SlashingEvents) > 5 {
		return "high"
	} else if len(vault.SlashingEvents) > 2 {
		return "medium"
	}
	return "low"
}