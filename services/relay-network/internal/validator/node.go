package validator

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/crosspay/relay-network/internal/config"
	"github.com/crosspay/relay-network/internal/p2p"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type ValidationRequest struct {
	ID           uint64    `json:"id"`
	PaymentID    uint64    `json:"payment_id"`
	MessageHash  string    `json:"message_hash"`
	RequiredSigs int       `json:"required_signatures"`
	Deadline     time.Time `json:"deadline"`
	IsHighValue  bool      `json:"is_high_value"`
}

type SignatureResult struct {
	RequestID uint64 `json:"request_id"`
	Signature string `json:"signature"`
	Signer    string `json:"signer"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type Node struct {
	privateKey     *ecdsa.PrivateKey
	address        common.Address
	config         *config.Config
	client         *ethclient.Client
	contract       *RelayValidatorContract
	
	pendingValidations map[uint64]*ValidationRequest
	signatures         map[uint64]map[string]string
	mutex              sync.RWMutex
	
	isRegistered bool
	stake        *big.Int
	status       string
}

type RelayValidatorContract struct {
	// Contract binding would go here
	address common.Address
}

func NewNode(privateKey *ecdsa.PrivateKey, cfg *config.Config) *Node {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	
	return &Node{
		privateKey:         privateKey,
		address:            address,
		config:             cfg,
		pendingValidations: make(map[uint64]*ValidationRequest),
		signatures:         make(map[uint64]map[string]string),
		status:             "starting",
	}
}

func (n *Node) Start(ctx context.Context) error {
	client, err := ethclient.Dial(n.config.RPCEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum client: %w", err)
	}
	n.client = client

	contractAddr := common.HexToAddress(n.config.ContractAddress)
	n.contract = &RelayValidatorContract{address: contractAddr}

	if err := n.checkRegistration(ctx); err != nil {
		log.Printf("Warning: Could not check registration status: %v", err)
	}

	n.status = "active"
	log.Printf("Validator node started with address: %s", n.address.Hex())
	
	go n.monitorValidationRequests(ctx)
	go n.performHealthCheck(ctx)
	
	return nil
}

func (n *Node) RegisterValidator(ctx context.Context, stakeAmount *big.Int) error {
	if n.isRegistered {
		return fmt.Errorf("validator already registered")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(n.privateKey, big.NewInt(n.config.ChainID))
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.Value = stakeAmount
	auth.GasLimit = uint64(300000)

	log.Printf("Registering validator with stake: %s ETH", stakeAmount.String())
	
	n.isRegistered = true
	n.stake = stakeAmount
	
	return nil
}

func (n *Node) ProcessValidationRequest(msg *p2p.ValidationMessage) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if _, exists := n.pendingValidations[msg.RequestID]; exists {
		return fmt.Errorf("validation request %d already exists", msg.RequestID)
	}

	// Convert ValidationMessage to ValidationRequest for internal processing
	req := &ValidationRequest{
		ID:          msg.RequestID,
		PaymentID:   msg.PaymentID,
		MessageHash: msg.MessageHash,
		RequiredSigs: 2, // Default required signatures
		Deadline:    msg.Timestamp.Add(5 * time.Minute), // Set reasonable deadline
		IsHighValue: false, // Can be determined based on amount if needed
	}

	n.pendingValidations[req.ID] = req
	n.signatures[req.ID] = make(map[string]string)

	log.Printf("Processing validation request %d for payment %d", req.ID, req.PaymentID)

	go n.signValidationRequest(req)
	
	return nil
}

func (n *Node) signValidationRequest(req *ValidationRequest) {
	messageHashBytes, err := hex.DecodeString(req.MessageHash[2:]) // Remove 0x prefix
	if err != nil {
		log.Printf("Failed to decode message hash for request %d: %v", req.ID, err)
		return
	}

	signature, err := crypto.Sign(messageHashBytes, n.privateKey)
	if err != nil {
		log.Printf("Failed to sign message for request %d: %v", req.ID, err)
		return
	}

	signatureHex := "0x" + hex.EncodeToString(signature)
	
	n.mutex.Lock()
	n.signatures[req.ID][n.address.Hex()] = signatureHex
	n.mutex.Unlock()

	log.Printf("Signed validation request %d with signature: %s", req.ID, signatureHex[:10]+"...")

	if err := n.submitSignatureToContract(req.ID, signature); err != nil {
		log.Printf("Failed to submit signature to contract: %v", err)
	}
}

func (n *Node) submitSignatureToContract(requestID uint64, signature []byte) error {
	auth, err := bind.NewKeyedTransactorWithChainID(n.privateKey, big.NewInt(n.config.ChainID))
	if err != nil {
		return fmt.Errorf("failed to create transactor: %w", err)
	}

	auth.GasLimit = uint64(200000)

	log.Printf("Submitting signature for request %d to contract", requestID)
	
	return nil
}

func (n *Node) GetValidationStatus(requestID uint64) (*ValidationRequest, bool) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	req, exists := n.pendingValidations[requestID]
	return req, exists
}

func (n *Node) GetSignatures(requestID uint64) map[string]string {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	sigs := make(map[string]string)
	for addr, sig := range n.signatures[requestID] {
		sigs[addr] = sig
	}
	return sigs
}

func (n *Node) monitorValidationRequests(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.cleanupExpiredRequests()
		}
	}
}

func (n *Node) cleanupExpiredRequests() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	now := time.Now()
	for id, req := range n.pendingValidations {
		if now.After(req.Deadline) {
			delete(n.pendingValidations, id)
			delete(n.signatures, id)
			log.Printf("Cleaned up expired validation request %d", id)
		}
	}
}

func (n *Node) performHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n.client != nil {
				if _, err := n.client.BlockNumber(ctx); err != nil {
					log.Printf("Health check failed: %v", err)
					n.status = "unhealthy"
				} else {
					n.status = "healthy"
				}
			}
		}
	}
}

func (n *Node) checkRegistration(ctx context.Context) error {
	log.Printf("Checking validator registration status for %s", n.address.Hex())
	
	n.isRegistered = false
	return nil
}

func (n *Node) GetAddress() string {
	return n.address.Hex()
}

func (n *Node) GetStatus() string {
	return n.status
}

func (n *Node) GetStake() string {
	if n.stake == nil {
		return "0"
	}
	return n.stake.String()
}

func (n *Node) IsRegistered() bool {
	return n.isRegistered
}

func (n *Node) GetPendingValidationCount() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	return len(n.pendingValidations)
}