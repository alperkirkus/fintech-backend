package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type userStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) store.UserStore {
	return &userStore{db: db}
}

func (s *userStore) Create(ctx context.Context, user *model.User) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Username, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (s *userStore) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (s *userStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return user, nil
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, store.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (s *userStore) Update(ctx context.Context, user *model.User) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET username = $1, email = $2, role = $3, updated_at = $4 WHERE id = $5`,
		user.Username, user.Email, user.Role, user.UpdatedAt, user.ID,
	)
	return err
}

func (s *userStore) List(ctx context.Context, limit, offset int) ([]*model.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, username, email, password_hash, role, created_at, updated_at
		 FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
