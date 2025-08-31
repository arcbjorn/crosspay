package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              int
	KeyPath           string
	ContractAddress   string
	RPCEndpoint       string
	ChainID           int64
	P2P               P2PConfig
	Validation        ValidationConfig
}

type P2PConfig struct {
	Port           int
	BootstrapPeers []string
	MaxPeers       int
}

type ValidationConfig struct {
	TimeoutSeconds    int
	MaxConcurrent     int
	SignatureRequired bool
}

func Load() *Config {
	return &Config{
		Port:            getEnvInt("PORT", 8080),
		KeyPath:         getEnv("KEY_PATH", "./validator.key"),
		ContractAddress: getEnv("CONTRACT_ADDRESS", ""),
		RPCEndpoint:     getEnv("RPC_ENDPOINT", "http://localhost:8545"),
		ChainID:         int64(getEnvInt("CHAIN_ID", 1337)),
		P2P: P2PConfig{
			Port:           getEnvInt("P2P_PORT", 9090),
			BootstrapPeers: strings.Split(getEnv("BOOTSTRAP_PEERS", ""), ","),
			MaxPeers:       getEnvInt("MAX_PEERS", 50),
		},
		Validation: ValidationConfig{
			TimeoutSeconds:    getEnvInt("VALIDATION_TIMEOUT", 300),
			MaxConcurrent:     getEnvInt("MAX_CONCURRENT_VALIDATIONS", 10),
			SignatureRequired: getEnv("SIGNATURE_REQUIRED", "true") == "true",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}