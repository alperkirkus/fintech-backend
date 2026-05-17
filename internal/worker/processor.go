package worker

import (
	"context"
	"fmt"
)

// Processor batch iş yüklerini worker pool üzerinden çalıştırır ve
// tüm sonuçları toplayarak döner.
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

// ProcessBatch verilen job'ları pool'a gönderir ve hepsinin tamamlanmasını
// bekler. Context iptal edilirse kısmen tamamlanmış sonuçları ve context
// hatasını birlikte döner.
func (p *Processor) ProcessBatch(ctx context.Context, jobs []Job) ([]Result, error) {
	if len(jobs) == 0 {
		return nil, nil
	}

	resultCh := make(chan Result, len(jobs))

	for i := range jobs {
		if err := p.sem.Acquire(ctx); err != nil {
			// Context iptal edildi, gönderilenlerin sonuçlarını topla ve dön.
			collected := p.drain(resultCh, i)
			return collected, fmt.Errorf("batch aborted after %d/%d jobs: %w", i, len(jobs), err)
		}

		job := jobs[i]
		originalCallback := job.Callback

		job.Callback = func(r Result) {
			p.sem.Release()
			resultCh <- r
			if originalCallback != nil {
				originalCallback(r)
			}
		}

		p.pool.Submit(job)
	}

	return p.drain(resultCh, len(jobs)), nil
}

// drain resultCh'dan n adet sonucu okur.
func (p *Processor) drain(ch <-chan Result, n int) []Result {
	results := make([]Result, 0, n)
	for range n {
		results = append(results, <-ch)
	}
	return results
}
