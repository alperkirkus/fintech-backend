package worker

import "context"

// Semaphore n eşzamanlı operasyona izin verir.
// Acquire context iptal edilirse hemen hata döner — servis katmanında
// timeout yönetimi için kullanılır.
type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, n)}
}

// Acquire bir slot talep eder. Context iptal olursa hata döner.
func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release bir slotu serbest bırakır. Her Acquire çağrısından sonra
// defer ile çağrılmalıdır.
func (s *Semaphore) Release() {
	<-s.ch
}

// TryAcquire bloklamadan slot almayı dener. Başarısız olursa false döner.
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// Available şu an boş olan slot sayısını döner.
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}
