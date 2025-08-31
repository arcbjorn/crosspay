package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

type StorageJob struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "upload", "receipt"
	Data        []byte                 `json:"data"`
	Filename    string                 `json:"filename"`
	PaymentID   uint64                 `json:"payment_id,omitempty"`
	Options     map[string]interface{} `json:"options"`
	CreatedAt   time.Time              `json:"created_at"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Status      string                 `json:"status"` // "pending", "processing", "completed", "failed"
	Error       string                 `json:"error,omitempty"`
	Result      *JobResult             `json:"result,omitempty"`
}

type JobResult struct {
	CID       string            `json:"cid"`
	Size      int64             `json:"size"`
	Cost      string            `json:"cost"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
}

type StorageQueue struct {
	jobs    map[string]*StorageJob
	pending chan *StorageJob
	mu      sync.RWMutex
	workers int
	ctx     context.Context
	cancel  context.CancelFunc
}

var queue *StorageQueue

func init() {
	queue = NewStorageQueue(3) // 3 workers
	queue.Start()
}

func NewStorageQueue(workers int) *StorageQueue {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &StorageQueue{
		jobs:    make(map[string]*StorageJob),
		pending: make(chan *StorageJob, 100),
		workers: workers,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (sq *StorageQueue) Start() {
	log.Printf("Starting storage queue with %d workers", sq.workers)
	
	for i := 0; i < sq.workers; i++ {
		go sq.worker(i)
	}
	
	// Start retry scheduler
	go sq.retryScheduler()
}

func (sq *StorageQueue) Stop() {
	log.Println("Stopping storage queue...")
	sq.cancel()
	close(sq.pending)
}

func (sq *StorageQueue) AddJob(job *StorageJob) error {
	job.ID = fmt.Sprintf("job_%d_%s", time.Now().UnixNano(), job.Type)
	job.CreatedAt = time.Now()
	job.Status = "pending"
	job.Attempts = 0
	job.MaxAttempts = 3

	sq.mu.Lock()
	sq.jobs[job.ID] = job
	sq.mu.Unlock()

	select {
	case sq.pending <- job:
		log.Printf("Job %s queued successfully", job.ID)
		return nil
	case <-sq.ctx.Done():
		return fmt.Errorf("queue is shutting down")
	default:
		return fmt.Errorf("queue is full")
	}
}

func (sq *StorageQueue) GetJob(jobID string) (*StorageJob, error) {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	
	job, exists := sq.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}
	
	return job, nil
}

func (sq *StorageQueue) worker(workerID int) {
	log.Printf("Storage worker %d started", workerID)
	
	for {
		select {
		case job := <-sq.pending:
			if job == nil {
				log.Printf("Worker %d stopping", workerID)
				return
			}
			sq.processJob(job, workerID)
			
		case <-sq.ctx.Done():
			log.Printf("Worker %d stopping due to context cancellation", workerID)
			return
		}
	}
}

func (sq *StorageQueue) processJob(job *StorageJob, workerID int) {
	log.Printf("Worker %d processing job %s (attempt %d)", workerID, job.ID, job.Attempts+1)
	
	sq.mu.Lock()
	job.Status = "processing"
	job.Attempts++
	sq.mu.Unlock()

	var result *JobResult
	var err error

	switch job.Type {
	case "upload":
		result, err = sq.processUploadJob(job)
	case "receipt":
		result, err = sq.processReceiptJob(job)
	default:
		err = fmt.Errorf("unknown job type: %s", job.Type)
	}

	sq.mu.Lock()
	defer sq.mu.Unlock()

	if err != nil {
		job.Error = err.Error()
		
		if job.Attempts >= job.MaxAttempts {
			job.Status = "failed"
			log.Printf("Job %s failed permanently after %d attempts: %v", job.ID, job.Attempts, err)
		} else {
			job.Status = "pending"
			log.Printf("Job %s failed (attempt %d/%d), will retry: %v", job.ID, job.Attempts, job.MaxAttempts, err)
			
			// Schedule retry
			go func() {
				delay := time.Duration(job.Attempts*job.Attempts) * time.Second // Exponential backoff
				time.Sleep(delay)
				
				select {
				case sq.pending <- job:
					log.Printf("Job %s requeued for retry", job.ID)
				case <-sq.ctx.Done():
					return
				}
			}()
		}
	} else {
		job.Status = "completed"
		job.Result = result
		log.Printf("Job %s completed successfully", job.ID)
	}
}

func (sq *StorageQueue) processUploadJob(job *StorageJob) (*JobResult, error) {
	cid, err := uploadToFilecoin(job.Data, job.Filename)
	if err != nil {
		return nil, err
	}

	cost := calculateStorageCost(int64(len(job.Data)))

	return &JobResult{
		CID:      cid,
		Size:     int64(len(job.Data)),
		Cost:     cost,
		Metadata: map[string]string{
			"filename":    job.Filename,
			"upload_type": "direct",
		},
		CreatedAt: time.Now(),
	}, nil
}

func (sq *StorageQueue) processReceiptJob(job *StorageJob) (*JobResult, error) {
	// Extract payment ID from options
	paymentIDFloat, ok := job.Options["payment_id"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid payment_id in job options")
	}
	paymentID := uint64(paymentIDFloat)

	format, _ := job.Options["format"].(string)
	if format == "" {
		format = "json"
	}

	language, _ := job.Options["language"].(string)
	if language == "" {
		language = "en"
	}

	// Fetch payment and generate receipt
	paymentData, err := fetchPaymentData(paymentID)
	if err != nil {
		return nil, err
	}

	receipt, err := generateReceipt(paymentData, format, language)
	if err != nil {
		return nil, err
	}

	// Convert to uploadable format
	var uploadData []byte
	var filename string

	switch format {
	case "pdf":
		uploadData, err = generatePDFReceipt(receipt)
		filename = fmt.Sprintf("receipt_%d.pdf", paymentID)
	default:
		uploadData, err = json.MarshalIndent(receipt, "", "  ")
		filename = fmt.Sprintf("receipt_%d.json", paymentID)
	}

	if err != nil {
		return nil, err
	}

	// Upload to Filecoin
	cid, err := uploadToFilecoin(uploadData, filename)
	if err != nil {
		return nil, err
	}

	receipt.CID = cid
	cost := calculateStorageCost(int64(len(uploadData)))

	return &JobResult{
		CID:  cid,
		Size: int64(len(uploadData)),
		Cost: cost,
		Metadata: map[string]string{
			"filename":   filename,
			"format":     format,
			"payment_id": strconv.FormatUint(paymentID, 10),
		},
		CreatedAt: time.Now(),
	}, nil
}

func (sq *StorageQueue) retryScheduler() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sq.checkFailedJobs()
		case <-sq.ctx.Done():
			return
		}
	}
}

func (sq *StorageQueue) checkFailedJobs() {
	sq.mu.RLock()
	defer sq.mu.RUnlock()

	failedCount := 0
	pendingCount := 0
	
	for _, job := range sq.jobs {
		switch job.Status {
		case "failed":
			failedCount++
		case "pending":
			pendingCount++
		}
	}

	if failedCount > 0 || pendingCount > 0 {
		log.Printf("Queue status: %d pending, %d failed jobs", pendingCount, failedCount)
	}
}