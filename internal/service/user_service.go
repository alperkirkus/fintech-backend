package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
	"github.com/alperkirkus/fintech-backend/internal/validator"
)

type UserService interface {
	Register(ctx context.Context, username, email, password string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
}

type userService struct {
	users store.UserStore
}

func NewUserService(users store.UserStore) UserService {
	return &userService{users: users}
}

func (s *userService) Register(ctx context.Context, username, email, password string) (*model.User, error) {
	if err := validator.Password(password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	u := &model.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := validator.User(u); err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}
