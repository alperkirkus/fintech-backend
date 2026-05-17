package service

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
	"github.com/alperkirkus/fintech-backend/internal/validator"
	"github.com/alperkirkus/fintech-backend/internal/worker"
)

type TransactionService interface {
	Deposit(ctx context.Context, toUserID uuid.UUID, amount decimal.Decimal) (*model.Transaction, error)
	Withdraw(ctx context.Context, fromUserID uuid.UUID, amount decimal.Decimal) (*model.Transaction, error)
	Transfer(ctx context.Context, fromUserID, toUserID uuid.UUID, amount decimal.Decimal) (*model.Transaction, error)
	Reverse(ctx context.Context, txID uuid.UUID) error
}

type transactionService struct {
	db   *sql.DB
	txs  store.TransactionStore
	sem  *worker.Semaphore
}

func NewTransactionService(db *sql.DB, txs store.TransactionStore, maxConcurrent int) TransactionService {
	return &transactionService{
		db:  db,
		txs: txs,
		sem: worker.NewSemaphore(maxConcurrent),
	}
}

func (s *transactionService) Deposit(ctx context.Context, toUserID uuid.UUID, amount decimal.Decimal) (*model.Transaction, error) {
	t := &model.Transaction{
		ToUserID: &toUserID,
		Amount:   amount,
		Type:     model.TypeDeposit,
	}
	if err := validator.Transaction(t); err != nil {
		return nil, err
	}

	if err := s.sem.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("acquire slot: %w", err)
	}
	defer s.sem.Release()

	return s.runInTx(ctx, func(tx *sql.Tx) (*model.Transaction, error) {
		if err := creditTx(ctx, tx, toUserID, amount); err != nil {
			return nil, err
		}
		return insertTransaction(ctx, tx, nil, &toUserID, amount, model.TypeDeposit)
	})
}

func (s *transactionService) Withdraw(ctx context.Context, fromUserID uuid.UUID, amount decimal.Decimal) (*model.Transaction, error) {
	t := &model.Transaction{
		FromUserID: &fromUserID,
		Amount:     amount,
		Type:       model.TypeWithdrawal,
	}
	if err := validator.Transaction(t); err != nil {
		return nil, err
	}

	if err := s.sem.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("acquire slot: %w", err)
	}
	defer s.sem.Release()

	return s.runInTx(ctx, func(tx *sql.Tx) (*model.Transaction, error) {
		balance, err := lockBalance(ctx, tx, fromUserID)
		if err != nil {
			return nil, err
		}
		if balance.LessThan(amount) {
			return nil, ErrInsufficientFunds
		}
		if err := debitTx(ctx, tx, fromUserID, amount); err != nil {
			return nil, err
		}
		return insertTransaction(ctx, tx, &fromUserID, nil, amount, model.TypeWithdrawal)
	})
}

func (s *transactionService) Transfer(ctx context.Context, fromUserID, toUserID uuid.UUID, amount decimal.Decimal) (*model.Transaction, error) {
	t := &model.Transaction{
		FromUserID: &fromUserID,
		ToUserID:   &toUserID,
		Amount:     amount,
		Type:       model.TypeTransfer,
	}
	if err := validator.Transaction(t); err != nil {
		return nil, err
	}

	if err := s.sem.Acquire(ctx); err != nil {
		return nil, fmt.Errorf("acquire slot: %w", err)
	}
	defer s.sem.Release()

	return s.runInTx(ctx, func(tx *sql.Tx) (*model.Transaction, error) {
		// Deadlock'u önlemek için satırları her zaman aynı sırada kilitle.
		ids := []uuid.UUID{fromUserID, toUserID}
		sort.Slice(ids, func(i, j int) bool {
			return ids[i].String() < ids[j].String()
		})

		balances := make(map[uuid.UUID]decimal.Decimal, 2)
		for _, id := range ids {
			bal, err := lockBalance(ctx, tx, id)
			if err != nil {
				return nil, err
			}
			balances[id] = bal
		}

		if balances[fromUserID].LessThan(amount) {
			return nil, ErrInsufficientFunds
		}

		if err := debitTx(ctx, tx, fromUserID, amount); err != nil {
			return nil, err
		}
		if err := creditTx(ctx, tx, toUserID, amount); err != nil {
			return nil, err
		}

		return insertTransaction(ctx, tx, &fromUserID, &toUserID, amount, model.TypeTransfer)
	})
}

func (s *transactionService) Reverse(ctx context.Context, txID uuid.UUID) error {
	original, err := s.txs.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}

	if !original.CanTransitionTo(model.StatusReversed) {
		return ErrInvalidTransition
	}

	if err := s.sem.Acquire(ctx); err != nil {
		return fmt.Errorf("acquire slot: %w", err)
	}
	defer s.sem.Release()

	_, err = s.runInTx(ctx, func(tx *sql.Tx) (*model.Transaction, error) {
		switch original.Type {
		case model.TypeTransfer:
			if err := debitTx(ctx, tx, *original.ToUserID, original.Amount); err != nil {
				return nil, err
			}
			if err := creditTx(ctx, tx, *original.FromUserID, original.Amount); err != nil {
				return nil, err
			}
		case model.TypeDeposit:
			bal, err := lockBalance(ctx, tx, *original.ToUserID)
			if err != nil {
				return nil, err
			}
			if bal.LessThan(original.Amount) {
				return nil, ErrInsufficientFunds
			}
			if err := debitTx(ctx, tx, *original.ToUserID, original.Amount); err != nil {
				return nil, err
			}
		case model.TypeWithdrawal:
			if err := creditTx(ctx, tx, *original.FromUserID, original.Amount); err != nil {
				return nil, err
			}
		}

		if _, err := tx.ExecContext(ctx,
			`UPDATE transactions SET status = $1 WHERE id = $2`,
			model.StatusReversed, txID,
		); err != nil {
			return nil, fmt.Errorf("update status: %w", err)
		}

		return nil, nil
	})

	return err
}

// runInTx bir DB transaction açar, fn'i çalıştırır; hata varsa rollback, yoksa commit yapar.
func (s *transactionService) runInTx(ctx context.Context, fn func(*sql.Tx) (*model.Transaction, error)) (*model.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck — no-op after Commit

	result, err := fn(tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return result, nil
}

// lockBalance SELECT FOR UPDATE ile balance satırını kilitler ve mevcut miktarı döner.
func lockBalance(ctx context.Context, tx *sql.Tx, userID uuid.UUID) (decimal.Decimal, error) {
	var amount decimal.Decimal
	err := tx.QueryRowContext(ctx,
		`SELECT amount FROM balances WHERE user_id = $1 FOR UPDATE`,
		userID,
	).Scan(&amount)
	if err == sql.ErrNoRows {
		return decimal.Zero, store.ErrNotFound
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("lock balance %s: %w", userID, err)
	}
	return amount, nil
}

func debitTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, amount decimal.Decimal) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE balances SET amount = amount - $1, last_updated_at = NOW() WHERE user_id = $2`,
		amount, userID,
	)
	return err
}

func creditTx(ctx context.Context, tx *sql.Tx, userID uuid.UUID, amount decimal.Decimal) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE balances SET amount = amount + $1, last_updated_at = NOW() WHERE user_id = $2`,
		amount, userID,
	)
	return err
}

func insertTransaction(
	ctx context.Context,
	tx *sql.Tx,
	fromID, toID *uuid.UUID,
	amount decimal.Decimal,
	txType model.TransactionType,
) (*model.Transaction, error) {
	record := &model.Transaction{
		ID:         uuid.New(),
		FromUserID: fromID,
		ToUserID:   toID,
		Amount:     amount,
		Type:       txType,
		Status:     model.StatusCompleted,
		CreatedAt:  time.Now().UTC(),
	}

	_, err := tx.ExecContext(ctx,
		`INSERT INTO transactions (id, from_user_id, to_user_id, amount, type, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		record.ID, record.FromUserID, record.ToUserID,
		record.Amount, record.Type, record.Status, record.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	return record, nil
}
