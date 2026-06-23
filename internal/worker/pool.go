package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID       uuid.UUID
	Task     func(ctx context.Context) error
	Callback func(Result)
}

type Result struct {
	JobID    uuid.UUID
	Err      error
	Duration time.Duration
}

type Stats struct {
	Processed int64
	Failed    int64
	Pending   int64
}

type Pool struct {
	jobs      chan Job
	waitGroup sync.WaitGroup

	processed atomic.Int64
	failed    atomic.Int64
	pending   atomic.Int64
}

func NewPool(ctx context.Context, workers, queueSize int) *Pool {
	p := &Pool{
		jobs: make(chan Job, queueSize),
	}

	for range workers {
		p.waitGroup.Add(1)
		go p.work(ctx)
	}

	return p
}

func (p *Pool) work(ctx context.Context) {
	defer p.waitGroup.Done()

	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			p.pending.Add(-1)

			start := time.Now()
			err := job.Task(ctx)
			duration := time.Since(start)

			if err != nil {
				p.failed.Add(1)
			} else {
				p.processed.Add(1)
			}

			if job.Callback != nil {
				job.Callback(Result{JobID: job.ID, Err: err, Duration: duration})
			}

		case <-ctx.Done():
			return
		}
	}
}

func (p *Pool) Submit(job Job) {
	p.pending.Add(1)
	p.jobs <- job
}

func (p *Pool) TrySubmit(job Job) bool {
	p.pending.Add(1)
	select {
	case p.jobs <- job:
		return true
	default:
		p.pending.Add(-1)
		return false
	}
}

func (p *Pool) Stop() {
	close(p.jobs)
	p.waitGroup.Wait()
}

func (p *Pool) Stats() Stats {
	return Stats{
		Processed: p.processed.Load(),
		Failed:    p.failed.Load(),
		Pending:   p.pending.Load(),
	}
}

func (p *Pool) ResetStats() {
	p.processed.Store(0)
	p.failed.Store(0)
	p.pending.Store(0)
}

func (p *Pool) QueueLen() int {
	return len(p.jobs)
}

func NewJob(task func(ctx context.Context) error) Job {
	return Job{
		ID:   uuid.New(),
		Task: task,
	}
}

type AtomicCounter struct {
	value atomic.Int64
}

func (c *AtomicCounter) Inc()        { c.value.Add(1) }
func (c *AtomicCounter) Dec()        { c.value.Add(-1) }
func (c *AtomicCounter) Add(n int64) { c.value.Add(n) }
func (c *AtomicCounter) Load() int64 { return c.value.Load() }
func (c *AtomicCounter) Reset()      { c.value.Store(0) }
