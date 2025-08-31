package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/crosspay/relay-network/internal/config"
	"github.com/crosspay/relay-network/internal/handlers"
	"github.com/crosspay/relay-network/internal/p2p"
	"github.com/crosspay/relay-network/internal/validator"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	cfg := config.Load()

	privateKey, err := loadOrGenerateKey(cfg.KeyPath)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	validatorNode := validator.NewNode(privateKey, cfg)
	p2pNetwork := p2p.NewNetwork(cfg.P2P, validatorNode)
	
	if err := p2pNetwork.Start(); err != nil {
		log.Fatalf("Failed to start P2P network: %v", err)
	}

	handler := handlers.NewHandler(validatorNode, p2pNetwork)
	
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /status", handler.Status)
	mux.HandleFunc("POST /validate", handler.RequestValidation)
	mux.HandleFunc("POST /sign", handler.SignMessage)
	mux.HandleFunc("GET /peers", handler.GetPeers)
	mux.HandleFunc("POST /register", handler.RegisterValidator)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting validator node on port %d", cfg.Port)
		log.Printf("Validator address: %s", crypto.PubkeyToAddress(privateKey.PublicKey).Hex())
		
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down validator node...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	p2pNetwork.Stop()
	log.Println("Validator node stopped")
}

func loadOrGenerateKey(keyPath string) (*ecdsa.PrivateKey, error) {
	if keyPath != "" {
		keyData, err := os.ReadFile(keyPath)
		if err == nil {
			keyHex := string(keyData)
			return crypto.HexToECDSA(keyHex)
		}
	}

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	if keyPath != "" {
		keyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))
		if err := os.WriteFile(keyPath, []byte(keyHex), 0600); err != nil {
			log.Printf("Warning: Could not save key to %s: %v", keyPath, err)
		}
	}

	return privateKey, nil
}