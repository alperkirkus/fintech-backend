package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type balanceStore struct {
	db *sql.DB
}

func NewBalanceStore(db *sql.DB) store.BalanceStore {
	return &balanceStore{db: db}
}

func (s *balanceStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Balance, error) {
	balance := &model.Balance{}
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, amount, last_updated_at FROM balances WHERE user_id = $1`,
		userID,
	).Scan(&balance.UserID, &balance.Amount, &balance.LastUpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	return balance, nil
}

func (s *balanceStore) Upsert(ctx context.Context, balance *model.Balance) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO balances (user_id, amount, last_updated_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET amount = $2, last_updated_at = $3`,
		balance.UserID, balance.Amount, balance.LastUpdatedAt,
	)
	return err
}
