package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type BalanceService interface {
	GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error)
	GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Transaction, error)
	GetAtTime(ctx context.Context, userID uuid.UUID, at time.Time) (decimal.Decimal, error)
}

type balanceService struct {
	balances     store.BalanceStore
	transactions store.TransactionStore
}

func NewBalanceService(balances store.BalanceStore, transactions store.TransactionStore) BalanceService {
	return &balanceService{
		balances:     balances,
		transactions: transactions,
	}
}

func (s *balanceService) GetBalance(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	balance, err := s.balances.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}

func (s *balanceService) GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Transaction, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	transactionList, err := s.transactions.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}
	if transactionList == nil {
		return []*model.Transaction{}, nil
	}
	return transactionList, nil
}

func (s *balanceService) GetAtTime(ctx context.Context, userID uuid.UUID, at time.Time) (decimal.Decimal, error) {
	amount, err := s.transactions.GetNetAmountByUserIDUntil(ctx, userID, at)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get balance at time: %w", err)
	}
	return amount, nil
}
