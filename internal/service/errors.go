package service

import "errors"

var (
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidTransition  = errors.New("invalid transaction state transition")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
)
