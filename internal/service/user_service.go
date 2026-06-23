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
	List(ctx context.Context, limit, offset int) ([]*model.User, error)
	Update(ctx context.Context, id uuid.UUID, username, email string) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
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
	user := &model.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := validator.User(user); err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (s *userService) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	users, err := s.users.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	if users == nil {
		return []*model.User{}, nil
	}
	return users, nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, username, email string) (*model.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	user.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	return user, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
