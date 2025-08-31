package pool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnectionPool(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	assert.NotNil(t, cp)
	assert.Equal(t, "ws://localhost:8545", cp.rpcEndpoint)
	assert.Equal(t, 5, cp.maxConns)
	assert.Equal(t, time.Minute, cp.idleTimeout)
	assert.NotNil(t, cp.connections)
	assert.NotNil(t, cp.activeConns)
	assert.False(t, cp.closed)
}

func TestConnectionPoolGet(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 2, time.Minute)
	ctx := context.Background()

	// First connection - should create new
	client1, err := cp.Get(ctx)
	assert.Error(t, err) // Will fail to connect to localhost:8545 in test
	assert.Nil(t, client1)
}

func TestConnectionPoolGetWhenClosed(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 2, time.Minute)
	cp.Close()
	
	ctx := context.Background()
	client, err := cp.Get(ctx)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection pool closed")
	assert.Nil(t, client)
}

func TestConnectionPoolStats(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	total, active, idle := cp.Stats()
	assert.Equal(t, 0, total)
	assert.Equal(t, 0, active)
	assert.Equal(t, 0, idle)
}

func TestConnectionPoolClose(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	assert.False(t, cp.closed)
	
	cp.Close()
	
	assert.True(t, cp.closed)
}

func TestConnectionPoolPutWhenClosed(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	cp.Close()
	
	// This should not panic - Put handles closed pool gracefully
	cp.Put(nil)
}

func TestConnectionPoolIsConnectionValid(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	now := time.Now()
	
	// Recent connection should be valid
	recentConn := &PooledConnection{
		LastUsed: now.Add(-30 * time.Second),
	}
	assert.True(t, cp.isConnectionValid(recentConn))
	
	// Old connection should be invalid
	oldConn := &PooledConnection{
		LastUsed: now.Add(-2 * time.Minute),
	}
	assert.False(t, cp.isConnectionValid(oldConn))
}

func TestConnectionPoolCleanup(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	now := time.Now()
	
	// Create expired connection
	expiredConn := &PooledConnection{
		LastUsed: now.Add(-2 * time.Minute),
	}
	
	// Create valid connection
	validConn := &PooledConnection{
		LastUsed: now.Add(-30 * time.Second),
	}
	
	// Add connections to pool (normally would have real clients)
	// This test verifies the cleanup logic without actual network connections
	
	// Test cleanup doesn't panic
	cp.Cleanup()
}

func TestConnectionPoolStartCleanup(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, 100*time.Millisecond) // Short timeout for test
	
	ctx, cancel := context.WithCancel(context.Background())
	
	// Start cleanup routine
	go cp.StartCleanup(ctx)
	
	// Let it run briefly
	time.Sleep(150 * time.Millisecond)
	
	// Cancel and verify no panic
	cancel()
	
	// Give time for cleanup routine to exit
	time.Sleep(50 * time.Millisecond)
}

func TestPooledConnectionFields(t *testing.T) {
	now := time.Now()
	
	conn := &PooledConnection{
		Client:    nil, // Would normally be *ethclient.Client
		CreatedAt: now,
		LastUsed:  now,
		InUse:     true,
	}
	
	assert.Nil(t, conn.Client)
	assert.Equal(t, now, conn.CreatedAt)
	assert.Equal(t, now, conn.LastUsed)
	assert.True(t, conn.InUse)
}

// Test concurrent access to stats
func TestConnectionPoolStatsConcurrent(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	done := make(chan bool)
	
	// Start multiple goroutines calling Stats
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cp.Stats()
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for concurrent stats calls")
		}
	}
}

// Test concurrent cleanup calls
func TestConnectionPoolCleanupConcurrent(t *testing.T) {
	cp := NewConnectionPool("ws://localhost:8545", 5, time.Minute)
	
	done := make(chan bool)
	
	// Start multiple goroutines calling Cleanup
	for i := 0; i < 5; i++ {
		go func() {
			cp.Cleanup()
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("Timeout waiting for concurrent cleanup calls")
		}
	}
}