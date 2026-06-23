package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/alperkirkus/fintech-backend/internal/model"
	"github.com/alperkirkus/fintech-backend/internal/store"
)

type Claims struct {
	UserID string     `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

type AuthService interface {
	Login(ctx context.Context, email, password string) (string, error)
	ValidateToken(token string) (*Claims, error)
	RequireRole(claims *Claims, role model.Role) error
	Refresh(claims *Claims) (string, error)
}

type authService struct {
	users     store.UserStore
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(users store.UserStore, jwtSecret string, tokenTTL time.Duration) AuthService {
	return &authService{
		users:     users,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	claims := &Claims{
		UserID: user.ID.String(),
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.tokenTTL)),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return token, nil
}

func (s *authService) ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrUnauthorized
	}

	return claims, nil
}

func (s *authService) RequireRole(claims *Claims, role model.Role) error {
	if claims.Role != role {
		return ErrUnauthorized
	}
	return nil
}

func (s *authService) Refresh(claims *Claims) (string, error) {
	newClaims := &Claims{
		UserID: claims.UserID,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(s.tokenTTL)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims).SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}
