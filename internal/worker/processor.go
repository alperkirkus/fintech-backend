package worker

import (
	"context"
	"fmt"
)

type Processor struct {
	pool *Pool
	sem  *Semaphore
}

func NewProcessor(pool *Pool, maxConcurrent int) *Processor {
	return &Processor{
		pool: pool,
		sem:  NewSemaphore(maxConcurrent),
	}
}

func (p *Processor) ProcessBatch(ctx context.Context, jobs []Job) ([]Result, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	resultChannel := make(chan Result, len(jobs))

	for i := range jobs {
		if err := p.sem.Acquire(ctx); err != nil {
			collected := p.drain(resultChannel, i)
			return collected, fmt.Errorf("batch aborted after %d/%d jobs: %w", i, len(jobs), err)
		}

		job := jobs[i]
		originalCallback := job.Callback

		job.Callback = func(r Result) {
			p.sem.Release()
			resultChannel <- r
			if originalCallback != nil {
				originalCallback(r)
			}
		}

		p.pool.Submit(job)
	}

	return p.drain(resultChannel, len(jobs)), nil
}

func (p *Processor) drain(resultChannel <-chan Result, n int) []Result {
	results := make([]Result, 0, n)
	for range n {
		results = append(results, <-resultChannel)
	}
	return results
}
