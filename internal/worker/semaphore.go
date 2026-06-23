package worker

import "context"

type Semaphore struct {
	slots chan struct{}
}

func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{slots: make(chan struct{}, capacity)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) Release() {
	<-s.slots
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Available() int {
	return cap(s.slots) - len(s.slots)
}
