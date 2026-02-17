package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Job representa una tarea a ejecutarse en el pool
type Job interface {
	Execute(ctx context.Context) error
	ID() string
	Name() string
}

// Result es el resultado de ejecutar un Job
type Result struct {
	JobID    string
	JobName  string
	Error    error
	Duration time.Duration
	Timestamp time.Time
}

// Pool es un worker pool para ejecutar jobs de forma asincrónica
type Pool struct {
	workers      int
	jobs         chan Job
	results      chan Result
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	totalJobs    int
	completedJobs int
	failedJobs   int
}

// NewPool crea un nuevo worker pool
func NewPool(ctx context.Context, workers int) *Pool {
	if workers < 1 {
		workers = 1
	}
	pctx, cancel := context.WithCancel(ctx)

	p := &Pool{
		workers: workers,
		jobs:    make(chan Job, workers*2),
		results: make(chan Result, workers*2),
		ctx:     pctx,
		cancel:  cancel,
	}

	// Iniciar workers
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Startup log
	log.Printf("🚀 Worker pool started with %d workers\n", workers)

	return p
}

// worker es la goroutine que procesa jobs
func (p *Pool) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("⏹️  Worker %d shutting down\n", id)
			return
		case job, ok := <-p.jobs:
			if !ok {
				log.Printf("⏹️  Worker %d: jobs channel closed\n", id)
				return
			}

			start := time.Now()
			err := job.Execute(p.ctx)
			duration := time.Since(start)

			result := Result{
				JobID:     job.ID(),
				JobName:   job.Name(),
				Error:     err,
				Duration:  duration,
				Timestamp: start,
			}

			// Actualizar stats
			p.mu.Lock()
			p.completedJobs++
			if err != nil {
				p.failedJobs++
				log.Printf("❌ Job %s (%s) failed after %dms: %v\n",
					job.ID(), job.Name(), duration.Milliseconds(), err)
			} else {
				log.Printf("✅ Job %s (%s) completed in %dms\n",
					job.ID(), job.Name(), duration.Milliseconds())
			}
			p.mu.Unlock()

			// Enviar resultado (non-blocking)
			select {
			case p.results <- result:
			case <-p.ctx.Done():
				return
			default:
				log.Printf("⚠️  Results channel full, dropping result for job %s\n", job.ID())
			}
		}
	}
}

// Submit envía un job al pool (non-blocking)
func (p *Pool) Submit(job Job) error {
	p.mu.Lock()
	p.totalJobs++
	p.mu.Unlock()

	select {
	case p.jobs <- job:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("pool is shutting down")
	default:
		return fmt.Errorf("job queue is full")
	}
}

// Results retorna el channel de resultados
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Wait espera a que todos los jobs se completen
func (p *Pool) Wait() {
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
}

// Shutdown detiene el pool gracefully
func (p *Pool) Shutdown(ctx context.Context) error {
	log.Println("🛑 Shutting down worker pool...")
	p.cancel()

	done := make(chan struct{})
	go func() {
		p.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("✓ Worker pool shut down successfully")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout")
	}
}

// Stats retorna estadísticas del pool
func (p *Pool) Stats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return map[string]interface{}{
		"total_jobs":     p.totalJobs,
		"completed_jobs": p.completedJobs,
		"failed_jobs":    p.failedJobs,
		"pending_jobs":   p.totalJobs - p.completedJobs,
	}
}
