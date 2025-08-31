package batch

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ValidationRequest struct {
	ID          uint64    `json:"id"`
	PaymentID   uint64    `json:"payment_id"`
	MessageHash string    `json:"message_hash"`
	Amount      uint64    `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	Callback    chan ValidationResult
}

type ValidationResult struct {
	RequestID uint64 `json:"request_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

type BatchProcessor struct {
	requestChan  chan *ValidationRequest
	batchSize    int
	batchTimeout time.Duration
	processor    func([]*ValidationRequest) []ValidationResult
	mutex        sync.RWMutex
	running      bool
}

func NewBatchProcessor(batchSize int, timeout time.Duration, processor func([]*ValidationRequest) []ValidationResult) *BatchProcessor {
	return &BatchProcessor{
		requestChan:  make(chan *ValidationRequest, 1000),
		batchSize:    batchSize,
		batchTimeout: timeout,
		processor:    processor,
	}
}

func (bp *BatchProcessor) Start(ctx context.Context) {
	bp.mutex.Lock()
	bp.running = true
	bp.mutex.Unlock()

	go bp.processBatches(ctx)
}

func (bp *BatchProcessor) Stop() {
	bp.mutex.Lock()
	bp.running = false
	bp.mutex.Unlock()

	close(bp.requestChan)
}

func (bp *BatchProcessor) Submit(req *ValidationRequest) error {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	if !bp.running {
		return fmt.Errorf("batch processor not running")
	}

	select {
	case bp.requestChan <- req:
		return nil
	default:
		return fmt.Errorf("batch processor queue full")
	}
}

func (bp *BatchProcessor) processBatches(ctx context.Context) {
	batch := make([]*ValidationRequest, 0, bp.batchSize)
	timer := time.NewTimer(bp.batchTimeout)

	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				bp.executeBatch(batch)
			}
			return

		case req, ok := <-bp.requestChan:
			if !ok {
				if len(batch) > 0 {
					bp.executeBatch(batch)
				}
				return
			}

			batch = append(batch, req)

			if len(batch) >= bp.batchSize {
				bp.executeBatch(batch)
				batch = batch[:0]
				timer.Reset(bp.batchTimeout)
			}

		case <-timer.C:
			if len(batch) > 0 {
				bp.executeBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(bp.batchTimeout)
		}
	}
}

func (bp *BatchProcessor) executeBatch(batch []*ValidationRequest) {
	if bp.processor == nil {
		return
	}

	results := bp.processor(batch)

	// Send results back through callbacks
	for i, req := range batch {
		if req.Callback != nil && i < len(results) {
			select {
			case req.Callback <- results[i]:
			default:
				// Channel full or closed, ignore
			}
		}
	}
}

func (bp *BatchProcessor) GetStats() (queueSize int, isRunning bool) {
	bp.mutex.RLock()
	defer bp.mutex.RUnlock()

	return len(bp.requestChan), bp.running
}