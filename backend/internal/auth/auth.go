package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const SessionLifetime = 24 * time.Hour

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrSessionExpired     = errors.New("authentication session expired")
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID         string
	UserID     string
	TokenHash  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	LastUsedAt time.Time
}

type Repository interface {
	CreateUser(ctx context.Context, user User) error
	GetUserByUsername(ctx context.Context, username string) (User, error)
	CreateSession(ctx context.Context, session Session) error
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	TouchSession(ctx context.Context, id string, lastUsedAt time.Time) error
	RevokeSession(ctx context.Context, id string, revokedAt time.Time) error
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	User      User
}

type Service struct {
	repository Repository
	now        func() time.Time
	newID      func() (string, error)
	newToken   func() (string, error)
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now, newID: newIdentifier, newToken: newToken}
}

func (s *Service) CreateUser(ctx context.Context, username string, password string) (User, error) {
	username, err := normalizeUsername(username)
	if err != nil {
		return User{}, err
	}
	if err := validatePassword(password); err != nil {
		return User{}, err
	}
	if s.repository == nil {
		return User{}, ErrInvalidCredentials
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}
	id, err := s.newID()
	if err != nil {
		return User{}, err
	}
	user := User{ID: id, Username: username, PasswordHash: passwordHash, CreatedAt: s.now().UTC()}
	if err := s.repository.CreateUser(ctx, user); err != nil {
		return User{}, err
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, username string, password string) (LoginResult, error) {
	username, err := normalizeUsername(username)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	if s.repository == nil {
		return LoginResult{}, ErrInvalidCredentials
	}
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil || !verifyPassword(user.PasswordHash, password) {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := s.newToken()
	if err != nil {
		return LoginResult{}, err
	}
	id, err := s.newID()
	if err != nil {
		return LoginResult{}, err
	}
	now := s.now().UTC()
	session := Session{ID: id, UserID: user.ID, TokenHash: hashToken(token), CreatedAt: now, ExpiresAt: now.Add(SessionLifetime), LastUsedAt: now}
	if err := s.repository.CreateSession(ctx, session); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token, ExpiresAt: session.ExpiresAt, User: user}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if token == "" || s.repository == nil {
		return User{}, ErrInvalidCredentials
	}
	session, err := s.repository.GetSessionByTokenHash(ctx, hashToken(token))
	if err != nil || session.RevokedAt != nil {
		return User{}, ErrInvalidCredentials
	}
	now := s.now().UTC()
	if !now.Before(session.ExpiresAt) {
		return User{}, ErrSessionExpired
	}
	if err := s.repository.TouchSession(ctx, session.ID, now); err != nil {
		return User{}, err
	}

	return User{ID: session.UserID}, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" || s.repository == nil {
		return ErrInvalidCredentials
	}
	session, err := s.repository.GetSessionByTokenHash(ctx, hashToken(token))
	if err != nil {
		return ErrInvalidCredentials
	}
	if session.RevokedAt != nil {
		return nil
	}

	return s.repository.RevokeSession(ctx, session.ID, s.now().UTC())
}

func normalizeUsername(username string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(username))
	if len(normalized) < 3 || len(normalized) > 64 {
		return "", ErrInvalidUsername
	}
	for _, character := range normalized {
		if !(character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' || character == '_') {
			return "", ErrInvalidUsername
		}
	}

	return normalized, nil
}

func validatePassword(password string) error {
	if len(password) < 12 || len(password) > 256 {
		return ErrInvalidPassword
	}

	return nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func newIdentifier() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate identifier: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}

func newToken() (string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", fmt.Errorf("generate authentication token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes[:]), nil
}
