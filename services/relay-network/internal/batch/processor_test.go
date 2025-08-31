package batch

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBatchProcessor(t *testing.T) {
	processor := func(reqs []*ValidationRequest) []ValidationResult {
		results := make([]ValidationResult, len(reqs))
		for i, req := range reqs {
			results[i] = ValidationResult{
				RequestID: req.ID,
				Success:   true,
			}
		}
		return results
	}

	bp := NewBatchProcessor(10, time.Second, processor)
	
	assert.NotNil(t, bp)
	assert.Equal(t, 10, bp.batchSize)
	assert.Equal(t, time.Second, bp.batchTimeout)
	assert.NotNil(t, bp.requestChan)
	assert.False(t, bp.running)
}

func TestBatchProcessorLifecycle(t *testing.T) {
	processor := func(reqs []*ValidationRequest) []ValidationResult {
		return make([]ValidationResult, len(reqs))
	}

	bp := NewBatchProcessor(10, time.Second, processor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start processor
	bp.Start(ctx)
	assert.True(t, bp.running)

	// Stop processor
	bp.Stop()
	assert.False(t, bp.running)
}

func TestBatchProcessorSubmit(t *testing.T) {
	processor := func(reqs []*ValidationRequest) []ValidationResult {
		return make([]ValidationResult, len(reqs))
	}

	bp := NewBatchProcessor(10, time.Second, processor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp.Start(ctx)
	defer bp.Stop()

	req := &ValidationRequest{
		ID:          1,
		PaymentID:   123,
		MessageHash: "test",
		Amount:      1000,
		Timestamp:   time.Now(),
	}

	err := bp.Submit(req)
	assert.NoError(t, err)
}

func TestBatchProcessorSubmitWhenNotRunning(t *testing.T) {
	processor := func(reqs []*ValidationRequest) []ValidationResult {
		return make([]ValidationResult, len(reqs))
	}

	bp := NewBatchProcessor(10, time.Second, processor)

	req := &ValidationRequest{
		ID:          1,
		PaymentID:   123,
		MessageHash: "test",
		Amount:      1000,
		Timestamp:   time.Now(),
	}

	err := bp.Submit(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestBatchProcessorExecuteBatch(t *testing.T) {
	var processedReqs []*ValidationRequest
	var mu sync.Mutex

	processor := func(reqs []*ValidationRequest) []ValidationResult {
		mu.Lock()
		processedReqs = append(processedReqs, reqs...)
		mu.Unlock()

		results := make([]ValidationResult, len(reqs))
		for i, req := range reqs {
			results[i] = ValidationResult{
				RequestID: req.ID,
				Success:   true,
			}
		}
		return results
	}

	bp := NewBatchProcessor(2, 100*time.Millisecond, processor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp.Start(ctx)
	defer bp.Stop()

	// Submit requests to trigger batch processing
	req1 := &ValidationRequest{ID: 1, PaymentID: 123, MessageHash: "test1", Amount: 1000, Timestamp: time.Now()}
	req2 := &ValidationRequest{ID: 2, PaymentID: 124, MessageHash: "test2", Amount: 2000, Timestamp: time.Now()}

	err := bp.Submit(req1)
	require.NoError(t, err)

	err = bp.Submit(req2)
	require.NoError(t, err)

	// Wait for batch processing
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	assert.Len(t, processedReqs, 2)
	assert.Equal(t, uint64(1), processedReqs[0].ID)
	assert.Equal(t, uint64(2), processedReqs[1].ID)
	mu.Unlock()
}

func TestBatchProcessorTimeoutTrigger(t *testing.T) {
	var processedReqs []*ValidationRequest
	var mu sync.Mutex

	processor := func(reqs []*ValidationRequest) []ValidationResult {
		mu.Lock()
		processedReqs = append(processedReqs, reqs...)
		mu.Unlock()

		results := make([]ValidationResult, len(reqs))
		for i, req := range reqs {
			results[i] = ValidationResult{
				RequestID: req.ID,
				Success:   true,
			}
		}
		return results
	}

	bp := NewBatchProcessor(10, 50*time.Millisecond, processor) // Small timeout
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp.Start(ctx)
	defer bp.Stop()

	// Submit single request (won't trigger size-based batching)
	req := &ValidationRequest{ID: 1, PaymentID: 123, MessageHash: "test", Amount: 1000, Timestamp: time.Now()}
	err := bp.Submit(req)
	require.NoError(t, err)

	// Wait for timeout-based processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, processedReqs, 1)
	assert.Equal(t, uint64(1), processedReqs[0].ID)
	mu.Unlock()
}

func TestBatchProcessorCallback(t *testing.T) {
	processor := func(reqs []*ValidationRequest) []ValidationResult {
		results := make([]ValidationResult, len(reqs))
		for i, req := range reqs {
			results[i] = ValidationResult{
				RequestID: req.ID,
				Success:   true,
			}
		}
		return results
	}

	bp := NewBatchProcessor(1, time.Second, processor)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp.Start(ctx)
	defer bp.Stop()

	callback := make(chan ValidationResult, 1)
	req := &ValidationRequest{
		ID:          1,
		PaymentID:   123,
		MessageHash: "test",
		Amount:      1000,
		Timestamp:   time.Now(),
		Callback:    callback,
	}

	err := bp.Submit(req)
	require.NoError(t, err)

	// Wait for result
	select {
	case result := <-callback:
		assert.Equal(t, uint64(1), result.RequestID)
		assert.True(t, result.Success)
	case <-time.After(time.Second):
		t.Fatal("Timeout waiting for callback")
	}
}

func TestBatchProcessorStats(t *testing.T) {
	processor := func(reqs []*ValidationRequest) []ValidationResult {
		return make([]ValidationResult, len(reqs))
	}

	bp := NewBatchProcessor(10, time.Second, processor)
	
	queueSize, isRunning := bp.GetStats()
	assert.Equal(t, 0, queueSize)
	assert.False(t, isRunning)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp.Start(ctx)
	defer bp.Stop()

	queueSize, isRunning = bp.GetStats()
	assert.Equal(t, 0, queueSize)
	assert.True(t, isRunning)
}

func TestBatchProcessorNilProcessor(t *testing.T) {
	bp := NewBatchProcessor(1, time.Second, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bp.Start(ctx)
	defer bp.Stop()

	req := &ValidationRequest{
		ID:          1,
		PaymentID:   123,
		MessageHash: "test",
		Amount:      1000,
		Timestamp:   time.Now(),
	}

	err := bp.Submit(req)
	assert.NoError(t, err)

	// Should not panic with nil processor
	time.Sleep(100 * time.Millisecond)
}

func TestBatchProcessorContextCancellation(t *testing.T) {
	var processedReqs []*ValidationRequest
	var mu sync.Mutex

	processor := func(reqs []*ValidationRequest) []ValidationResult {
		mu.Lock()
		processedReqs = append(processedReqs, reqs...)
		mu.Unlock()
		return make([]ValidationResult, len(reqs))
	}

	bp := NewBatchProcessor(10, time.Second, processor)
	ctx, cancel := context.WithCancel(context.Background())

	bp.Start(ctx)

	req := &ValidationRequest{ID: 1, PaymentID: 123, MessageHash: "test", Amount: 1000, Timestamp: time.Now()}
	err := bp.Submit(req)
	require.NoError(t, err)

	// Cancel context to trigger shutdown
	cancel()

	// Wait for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Len(t, processedReqs, 1) // Should process pending requests on shutdown
	mu.Unlock()
}