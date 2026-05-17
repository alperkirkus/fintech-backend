package validator

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/alperkirkus/fintech-backend/internal/model"
)

var (
	ErrUsernameRequired  = errors.New("username is required")
	ErrUsernameTooShort  = errors.New("username must be at least 3 characters")
	ErrUsernameTooLong   = errors.New("username must be at most 50 characters")
	ErrEmailRequired     = errors.New("email is required")
	ErrEmailInvalid      = errors.New("email is invalid")
	ErrPasswordRequired  = errors.New("password is required")
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrRoleInvalid       = errors.New("role must be 'user' or 'admin'")
	ErrAmountNonPositive = errors.New("amount must be greater than zero")
	ErrTypeInvalid       = errors.New("invalid transaction type")
	ErrFromUserRequired  = errors.New("from_user_id is required for withdrawal and transfer")
	ErrToUserRequired    = errors.New("to_user_id is required for deposit and transfer")
)

func User(u *model.User) error {
	if strings.TrimSpace(u.Username) == "" {
		return ErrUsernameRequired
	}
	if len(u.Username) < 3 {
		return ErrUsernameTooShort
	}
	if len(u.Username) > 50 {
		return ErrUsernameTooLong
	}
	if strings.TrimSpace(u.Email) == "" {
		return ErrEmailRequired
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return ErrEmailInvalid
	}
	if !u.Role.IsValid() {
		return ErrRoleInvalid
	}
	return nil
}

func Password(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrPasswordRequired
	}
	if len(password) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}

func Transaction(t *model.Transaction) error {
	if !t.Type.IsValid() {
		return ErrTypeInvalid
	}
	if t.Amount.LessThanOrEqual(decimal.Zero) {
		return ErrAmountNonPositive
	}
	switch t.Type {
	case model.TypeWithdrawal, model.TypeTransfer:
		if t.FromUserID == nil {
			return ErrFromUserRequired
		}
	}
	switch t.Type {
	case model.TypeDeposit, model.TypeTransfer:
		if t.ToUserID == nil {
			return ErrToUserRequired
		}
	}
	return nil
}
