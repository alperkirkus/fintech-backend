package store

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/alperkirkus/fintech-backend/internal/model"
)

var ErrNotFound = errors.New("record not found")

type UserStore interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
}

type TransactionStore interface {
	Create(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.TransactionStatus) error
}

type BalanceStore interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Balance, error)
	Upsert(ctx context.Context, balance *model.Balance) error
}

type AuditLogStore interface {
	Create(ctx context.Context, log *model.AuditLog) error
	GetByEntityID(ctx context.Context, entityType string, entityID uuid.UUID) ([]*model.AuditLog, error)
}
