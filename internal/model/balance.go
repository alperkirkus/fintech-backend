package model

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Balance struct {
	mu            sync.RWMutex
	UserID        uuid.UUID       `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	LastUpdatedAt time.Time       `json:"last_updated_at"`
}

func (b *Balance) Get() decimal.Decimal {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount
}

func (b *Balance) Set(amount decimal.Decimal) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Amount = amount
	b.LastUpdatedAt = time.Now().UTC()
}

func (b *Balance) IsNegative() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount.IsNegative()
}

func (b *Balance) HasSufficientFunds(required decimal.Decimal) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Amount.GreaterThanOrEqual(required)
}
