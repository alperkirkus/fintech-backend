package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string
type TransactionStatus string

const (
	TypeDeposit    TransactionType = "deposit"
	TypeWithdrawal TransactionType = "withdrawal"
	TypeTransfer   TransactionType = "transfer"
)

const (
	StatusPending   TransactionStatus = "pending"
	StatusCompleted TransactionStatus = "completed"
	StatusFailed    TransactionStatus = "failed"
	StatusReversed  TransactionStatus = "reversed"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case TypeDeposit, TypeWithdrawal, TypeTransfer:
		return true
	}
	return false
}

func (s TransactionStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusCompleted, StatusFailed, StatusReversed:
		return true
	}
	return false
}

var allowedTransitions = map[TransactionStatus]map[TransactionStatus]bool{
	StatusPending:   {StatusCompleted: true, StatusFailed: true},
	StatusCompleted: {StatusReversed: true},
	StatusFailed:    {},
	StatusReversed:  {},
}

type Transaction struct {
	ID         uuid.UUID         `json:"id"`
	FromUserID *uuid.UUID        `json:"from_user_id,omitempty"`
	ToUserID   *uuid.UUID        `json:"to_user_id,omitempty"`
	Amount     decimal.Decimal   `json:"amount"`
	Type       TransactionType   `json:"type"`
	Status     TransactionStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
}

func (t *Transaction) CanTransitionTo(next TransactionStatus) bool {
	targets, ok := allowedTransitions[t.Status]
	if !ok {
		return false
	}
	return targets[next]
}

func (t *Transaction) IsTerminal() bool {
	return t.Status == StatusFailed || t.Status == StatusReversed
}
