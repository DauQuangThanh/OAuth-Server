package workers

import (
	"context"
	"sync"
	"time"
)

// Task represents a unit of work to be executed
type Task struct {
	ID       string
	Handler  func(ctx context.Context) error
	Priority int
	Created  time.Time
}

// WorkerPool manages a pool of workers for concurrent task execution
type WorkerPool struct {
	workers      int
	taskQueue    chan *Task
	resultChan   chan *TaskResult
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	stats        *PoolStats
	stopped      bool
	taskClosed   bool
	resultClosed bool
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID   string
	Error    error
	Duration time.Duration
}

// PoolStats tracks worker pool statistics
type PoolStats struct {
	TasksCompleted int64
	TasksFailed    int64
	TotalDuration  time.Duration
	ActiveWorkers  int64
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workers int, bufferSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workers:    workers,
		taskQueue:  make(chan *Task, bufferSize),
		resultChan: make(chan *TaskResult, bufferSize),
		ctx:        ctx,
		cancel:     cancel,
		stats:      &PoolStats{},
	}
}

// Start begins the worker pool operation
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}

	// Start result processor
	go wp.processResults()
}

// worker is the main worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	wp.mu.Lock()
	wp.stats.ActiveWorkers++
	wp.mu.Unlock()

	defer func() {
		wp.mu.Lock()
		wp.stats.ActiveWorkers--
		wp.mu.Unlock()
	}()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task := <-wp.taskQueue:
			if task == nil {
				return
			}

			start := time.Now()
			err := task.Handler(wp.ctx)
			duration := time.Since(start)

			result := &TaskResult{
				TaskID:   task.ID,
				Error:    err,
				Duration: duration,
			}

			select {
			case wp.resultChan <- result:
			case <-wp.ctx.Done():
				return
			}
		}
	}
}

// processResults processes task results and updates statistics
func (wp *WorkerPool) processResults() {
	for {
		select {
		case <-wp.ctx.Done():
			return
		case result := <-wp.resultChan:
			wp.mu.Lock()
			if result.Error != nil {
				wp.stats.TasksFailed++
			} else {
				wp.stats.TasksCompleted++
			}
			wp.stats.TotalDuration += result.Duration
			wp.mu.Unlock()
		}
	}
}

// SubmitTask submits a task to the worker pool
func (wp *WorkerPool) SubmitTask(task *Task) error {
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return ErrPoolClosed
	default:
		return ErrQueueFull
	}
}

// SubmitTaskWithTimeout submits a task with a timeout
func (wp *WorkerPool) SubmitTaskWithTimeout(task *Task, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(wp.ctx, timeout)
	defer cancel()

	select {
	case wp.taskQueue <- task:
		return nil
	case <-ctx.Done():
		return ErrSubmissionTimeout
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	if wp.stopped {
		wp.mu.Unlock()
		return
	}
	wp.stopped = true
	wp.mu.Unlock()

	wp.cancel()

	// Close task queue if not already closed
	wp.mu.Lock()
	if !wp.taskClosed {
		close(wp.taskQueue)
		wp.taskClosed = true
	}
	wp.mu.Unlock()

	wp.wg.Wait()

	// Close result channel if not already closed
	wp.mu.Lock()
	if !wp.resultClosed {
		close(wp.resultChan)
		wp.resultClosed = true
	}
	wp.mu.Unlock()
}

// GetStats returns current pool statistics
func (wp *WorkerPool) GetStats() PoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()
	return *wp.stats
}

// GetStatsMap returns statistics as a map for JSON serialization
func (wp *WorkerPool) GetStatsMap() map[string]interface{} {
	stats := wp.GetStats()

	avgDuration := time.Duration(0)
	totalTasks := stats.TasksCompleted + stats.TasksFailed
	if totalTasks > 0 {
		avgDuration = stats.TotalDuration / time.Duration(totalTasks)
	}

	return map[string]interface{}{
		"active_workers":  stats.ActiveWorkers,
		"tasks_completed": stats.TasksCompleted,
		"tasks_failed":    stats.TasksFailed,
		"total_tasks":     totalTasks,
		"success_rate":    float64(stats.TasksCompleted) / float64(totalTasks+1),
		"avg_duration_ms": avgDuration.Milliseconds(),
		"queue_size":      len(wp.taskQueue),
		"queue_capacity":  cap(wp.taskQueue),
	}
}

// BatchTaskProcessor processes multiple tasks concurrently
type BatchTaskProcessor struct {
	pool       *WorkerPool
	batchSize  int
	maxWorkers int
}

// NewBatchTaskProcessor creates a new batch task processor
func NewBatchTaskProcessor(batchSize, maxWorkers int) *BatchTaskProcessor {
	pool := NewWorkerPool(maxWorkers, batchSize*2)
	pool.Start()

	return &BatchTaskProcessor{
		pool:       pool,
		batchSize:  batchSize,
		maxWorkers: maxWorkers,
	}
}

// ProcessBatch processes a batch of tasks
func (bp *BatchTaskProcessor) ProcessBatch(ctx context.Context, tasks []*Task) error {
	if len(tasks) == 0 {
		return nil
	}

	// Submit all tasks
	for _, task := range tasks {
		if err := bp.pool.SubmitTask(task); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the batch processor
func (bp *BatchTaskProcessor) Close() {
	bp.pool.Stop()
}

// GetStats returns batch processor statistics
func (bp *BatchTaskProcessor) GetStats() map[string]interface{} {
	return bp.pool.GetStatsMap()
}

// Error definitions
var (
	ErrPoolClosed        = &WorkerError{Message: "worker pool is closed"}
	ErrQueueFull         = &WorkerError{Message: "task queue is full"}
	ErrSubmissionTimeout = &WorkerError{Message: "task submission timeout"}
)

// WorkerError represents a worker-related error
type WorkerError struct {
	Message string
}

func (e *WorkerError) Error() string {
	return e.Message
}
