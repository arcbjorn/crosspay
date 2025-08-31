package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/crosspay/relay-network/internal/config"
)

type ValidationMessage struct {
	Type        string      `json:"type"`
	RequestID   uint64      `json:"request_id"`
	PaymentID   uint64      `json:"payment_id"`
	MessageHash string      `json:"message_hash"`
	Signature   string      `json:"signature,omitempty"`
	Signer      string      `json:"signer,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

type Peer struct {
	Address    string    `json:"address"`
	PublicKey  string    `json:"public_key"`
	LastSeen   time.Time `json:"last_seen"`
	Connection net.Conn  `json:"-"`
	IsActive   bool      `json:"is_active"`
}

type ValidatorNode interface {
	ProcessValidationRequest(req *ValidationMessage) error
	GetAddress() string
	GetStatus() string
}

type Network struct {
	config        config.P2PConfig
	validator     ValidatorNode
	peers         map[string]*Peer
	listener      net.Listener
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	messageQueue  chan *ValidationMessage
	isRunning     bool
}

func NewNetwork(cfg config.P2PConfig, validator ValidatorNode) *Network {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Network{
		config:       cfg,
		validator:    validator,
		peers:        make(map[string]*Peer),
		ctx:          ctx,
		cancel:       cancel,
		messageQueue: make(chan *ValidationMessage, 100),
	}
}

func (n *Network) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", n.config.Port))
	if err != nil {
		return fmt.Errorf("failed to start P2P listener: %w", err)
	}
	
	n.listener = listener
	n.isRunning = true
	
	log.Printf("P2P network listening on port %d", n.config.Port)

	go n.acceptConnections()
	go n.processMessages()
	go n.connectToBootstrapPeers()
	go n.maintainPeers()

	return nil
}

func (n *Network) Stop() {
	n.isRunning = false
	n.cancel()
	
	if n.listener != nil {
		n.listener.Close()
	}

	n.mutex.Lock()
	for _, peer := range n.peers {
		if peer.Connection != nil {
			peer.Connection.Close()
		}
	}
	n.mutex.Unlock()

	close(n.messageQueue)
	log.Println("P2P network stopped")
}

func (n *Network) acceptConnections() {
	for n.isRunning {
		conn, err := n.listener.Accept()
		if err != nil {
			if n.isRunning {
				log.Printf("Failed to accept connection: %v", err)
			}
			continue
		}

		go n.handleConnection(conn)
	}
}

func (n *Network) handleConnection(conn net.Conn) {
	defer conn.Close()

	peerAddr := conn.RemoteAddr().String()
	log.Printf("New peer connection from %s", peerAddr)

	peer := &Peer{
		Address:    peerAddr,
		LastSeen:   time.Now(),
		Connection: conn,
		IsActive:   true,
	}

	n.mutex.Lock()
	n.peers[peerAddr] = peer
	n.mutex.Unlock()

	defer func() {
		n.mutex.Lock()
		delete(n.peers, peerAddr)
		n.mutex.Unlock()
		log.Printf("Peer %s disconnected", peerAddr)
	}()

	decoder := json.NewDecoder(conn)
	for {
		var msg ValidationMessage
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Failed to decode message from peer %s: %v", peerAddr, err)
			break
		}

		peer.LastSeen = time.Now()
		n.messageQueue <- &msg
	}
}

func (n *Network) processMessages() {
	for msg := range n.messageQueue {
		if err := n.handleValidationMessage(msg); err != nil {
			log.Printf("Failed to handle validation message: %v", err)
		}
	}
}

func (n *Network) handleValidationMessage(msg *ValidationMessage) error {
	switch msg.Type {
	case "validation_request":
		req := &ValidationMessage{
			Type:        "validation_request",
			RequestID:   msg.RequestID,
			PaymentID:   msg.PaymentID,
			MessageHash: msg.MessageHash,
			Timestamp:   msg.Timestamp,
		}
		return n.validator.ProcessValidationRequest(req)
		
	case "signature_share":
		log.Printf("Received signature share for request %d from %s", msg.RequestID, msg.Signer)
		return n.aggregateSignature(msg)
		
	case "validation_complete":
		log.Printf("Validation %d completed", msg.RequestID)
		return nil
		
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

func (n *Network) aggregateSignature(msg *ValidationMessage) error {
	log.Printf("Aggregating signature for request %d", msg.RequestID)
	return nil
}

func (n *Network) BroadcastValidationRequest(req *ValidationMessage) error {
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal validation request: %w", err)
	}

	n.mutex.RLock()
	defer n.mutex.RUnlock()

	successCount := 0
	for addr, peer := range n.peers {
		if !peer.IsActive || peer.Connection == nil {
			continue
		}

		if _, err := peer.Connection.Write(data); err != nil {
			log.Printf("Failed to send message to peer %s: %v", addr, err)
			peer.IsActive = false
		} else {
			successCount++
		}
	}

	log.Printf("Broadcasted validation request %d to %d peers", req.RequestID, successCount)
	return nil
}

func (n *Network) BroadcastSignature(requestID uint64, signature string) error {
	msg := &ValidationMessage{
		Type:      "signature_share",
		RequestID: requestID,
		Signature: signature,
		Signer:    n.validator.GetAddress(),
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal signature: %w", err)
	}

	n.mutex.RLock()
	defer n.mutex.RUnlock()

	for addr, peer := range n.peers {
		if !peer.IsActive || peer.Connection == nil {
			continue
		}

		if _, err := peer.Connection.Write(data); err != nil {
			log.Printf("Failed to send signature to peer %s: %v", addr, err)
			peer.IsActive = false
		}
	}

	return nil
}

func (n *Network) connectToBootstrapPeers() {
	for _, peerAddr := range n.config.BootstrapPeers {
		if peerAddr == "" {
			continue
		}

		go func(addr string) {
			for n.isRunning {
				if err := n.connectToPeer(addr); err != nil {
					log.Printf("Failed to connect to bootstrap peer %s: %v", addr, err)
					time.Sleep(30 * time.Second)
					continue
				}
				break
			}
		}(peerAddr)
	}
}

func (n *Network) connectToPeer(peerAddr string) error {
	conn, err := net.DialTimeout("tcp", peerAddr, 10*time.Second)
	if err != nil {
		return err
	}

	go n.handleConnection(conn)
	return nil
}

func (n *Network) maintainPeers() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			n.cleanupInactivePeers()
		}
	}
}

func (n *Network) cleanupInactivePeers() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for addr, peer := range n.peers {
		if peer.LastSeen.Before(cutoff) {
			if peer.Connection != nil {
				peer.Connection.Close()
			}
			delete(n.peers, addr)
			log.Printf("Removed inactive peer %s", addr)
		}
	}
}

func (n *Network) GetPeers() []*Peer {
	n.mutex.RLock()
	defer n.mutex.RUnlock()

	peers := make([]*Peer, 0, len(n.peers))
	for _, peer := range n.peers {
		peerCopy := &Peer{
			Address:  peer.Address,
			LastSeen: peer.LastSeen,
			IsActive: peer.IsActive,
		}
		peers = append(peers, peerCopy)
	}

	return peers
}

func (n *Network) GetPeerCount() int {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	
	return len(n.peers)
}

func (n *Network) IsRunning() bool {
	return n.isRunning
}