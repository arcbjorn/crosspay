package pool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type ConnectionPool struct {
	rpcEndpoint  string
	maxConns     int
	idleTimeout  time.Duration
	connections  chan *PooledConnection
	activeConns  map[*ethclient.Client]*PooledConnection
	mutex        sync.RWMutex
	closed       bool
}

type PooledConnection struct {
	Client    *ethclient.Client
	CreatedAt time.Time
	LastUsed  time.Time
	InUse     bool
}

func NewConnectionPool(rpcEndpoint string, maxConns int, idleTimeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		rpcEndpoint:  rpcEndpoint,
		maxConns:     maxConns,
		idleTimeout:  idleTimeout,
		connections:  make(chan *PooledConnection, maxConns),
		activeConns:  make(map[*ethclient.Client]*PooledConnection),
	}
}

func (cp *ConnectionPool) Get(ctx context.Context) (*ethclient.Client, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if cp.closed {
		return nil, fmt.Errorf("connection pool closed")
	}

	// Try to get an existing connection
	select {
	case conn := <-cp.connections:
		if cp.isConnectionValid(conn) {
			conn.InUse = true
			conn.LastUsed = time.Now()
			cp.activeConns[conn.Client] = conn
			return conn.Client, nil
		}
		// Connection expired, close it
		conn.Client.Close()
	default:
		// No available connections
	}

	// Create new connection if under limit
	if len(cp.activeConns) < cp.maxConns {
		client, err := ethclient.DialContext(ctx, cp.rpcEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}

		conn := &PooledConnection{
			Client:    client,
			CreatedAt: time.Now(),
			LastUsed:  time.Now(),
			InUse:     true,
		}

		cp.activeConns[client] = conn
		return client, nil
	}

	return nil, fmt.Errorf("connection pool exhausted")
}

func (cp *ConnectionPool) Put(client *ethclient.Client) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	if cp.closed {
		client.Close()
		return
	}

	conn, exists := cp.activeConns[client]
	if !exists {
		client.Close()
		return
	}

	conn.InUse = false
	conn.LastUsed = time.Now()
	delete(cp.activeConns, client)

	if cp.isConnectionValid(conn) {
		select {
		case cp.connections <- conn:
			// Successfully returned to pool
		default:
			// Pool full, close connection
			conn.Client.Close()
		}
	} else {
		conn.Client.Close()
	}
}

func (cp *ConnectionPool) Close() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	cp.closed = true

	// Close all pooled connections
	close(cp.connections)
	for conn := range cp.connections {
		conn.Client.Close()
	}

	// Close all active connections
	for _, conn := range cp.activeConns {
		conn.Client.Close()
	}
}

func (cp *ConnectionPool) isConnectionValid(conn *PooledConnection) bool {
	return time.Since(conn.LastUsed) < cp.idleTimeout
}

func (cp *ConnectionPool) Stats() (total, active, idle int) {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()

	active = len(cp.activeConns)
	idle = len(cp.connections)
	total = active + idle

	return total, active, idle
}

func (cp *ConnectionPool) Cleanup() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Remove expired connections from pool
	var validConns []*PooledConnection
	
	for {
		select {
		case conn := <-cp.connections:
			if cp.isConnectionValid(conn) {
				validConns = append(validConns, conn)
			} else {
				conn.Client.Close()
			}
		default:
			// No more connections in pool
			goto done
		}
	}

done:
	// Put valid connections back
	for _, conn := range validConns {
		select {
		case cp.connections <- conn:
		default:
			conn.Client.Close()
		}
	}
}

// Start cleanup goroutine
func (cp *ConnectionPool) StartCleanup(ctx context.Context) {
	ticker := time.NewTicker(cp.idleTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cp.Cleanup()
		}
	}
}