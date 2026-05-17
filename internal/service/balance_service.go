package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error)
	GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Transaction, error)
}

type balanceService struct {
	balances store.BalanceStore
	txs      store.TransactionStore
}

func NewBalanceService(balances store.BalanceStore, txs store.TransactionStore) BalanceService {
	return &balanceService{
		balances: balances,
		txs:      txs,
	}
}

func (s *balanceService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	b, err := s.balances.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	return b, nil
}

// GetHistory kullanıcının işlem geçmişini sayfalı olarak döner.
// Tarihsel bakiye optimizasyonu: işlemler DB'de zaten DESC sıralı index ile tutulduğundan
// ek hesaplama gerekmez.
func (s *balanceService) GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	txs, err := s.txs.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	return txs, nil
}
