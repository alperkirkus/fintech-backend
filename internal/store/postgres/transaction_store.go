package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type transactionStore struct {
	db *sql.DB
}

func NewTransactionStore(db *sql.DB) store.TransactionStore {
	return &transactionStore{db: db}
}

func (s *transactionStore) Create(ctx context.Context, transaction *model.Transaction) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO transactions (id, from_user_id, to_user_id, amount, type, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		transaction.ID, transaction.FromUserID, transaction.ToUserID, transaction.Amount, transaction.Type, transaction.Status, transaction.CreatedAt,
	)
	return err
}

func (s *transactionStore) GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	transaction := &model.Transaction{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		 FROM transactions WHERE id = $1`,
		id,
	).Scan(&transaction.ID, &transaction.FromUserID, &transaction.ToUserID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	return transaction, nil
}

func (s *transactionStore) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Transaction, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, from_user_id, to_user_id, amount, type, status, created_at
		 FROM transactions
		 WHERE from_user_id = $1 OR to_user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*model.Transaction
	for rows.Next() {
		transaction := &model.Transaction{}
		if err := rows.Scan(&transaction.ID, &transaction.FromUserID, &transaction.ToUserID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	return transactions, rows.Err()
}

func (s *transactionStore) UpdateStatus(ctx context.Context, id uuid.UUID, status model.TransactionStatus) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE transactions SET status = $1 WHERE id = $2`,
		status, id,
	)
	return err
}

func (s *transactionStore) GetNetAmountByUserIDUntil(ctx context.Context, userID uuid.UUID, until time.Time) (decimal.Decimal, error) {
	var net decimal.Decimal
	err := s.db.QueryRowContext(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN to_user_id = $1 THEN amount ELSE 0 END), 0) -
			COALESCE(SUM(CASE WHEN from_user_id = $1 THEN amount ELSE 0 END), 0)
		 FROM transactions
		 WHERE (from_user_id = $1 OR to_user_id = $1)
		   AND status = 'completed'
		   AND created_at <= $2`,
		userID, until,
	).Scan(&net)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get net amount: %w", err)
	}
	return net, nil
}
