package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrMissingDatabasePath = errors.New("missing authentication database path")
	ErrUserNotFound        = errors.New("user not found")
	ErrUsernameTaken       = errors.New("username already exists")
	ErrSessionNotFound     = errors.New("authentication session not found")
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(databasePath string) (*SQLiteRepository, error) {
	if databasePath == "" {
		return nil, ErrMissingDatabasePath
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open authentication database: %w", err)
	}
	repository := &SQLiteRepository{db: db}
	if err := repository.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return repository, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteRepository) CreateUser(ctx context.Context, user User) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, username, password_hash, created_at)
		VALUES (?, ?, ?, ?)
	`, user.ID, user.Username, user.PasswordHash, user.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		if sqliteConstraint(err) {
			return ErrUsernameTaken
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, created_at
		FROM users
		WHERE username = ?
	`, username)
	var user User
	var createdAt string
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user: %w", err)
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return User{}, fmt.Errorf("parse user creation time: %w", err)
	}
	user.CreatedAt = parsedCreatedAt

	return user, nil
}

func (r *SQLiteRepository) CreateSession(ctx context.Context, session Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO auth_sessions (id, user_id, token_hash, created_at, expires_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.TokenHash, session.CreatedAt.UTC().Format(time.RFC3339Nano), session.ExpiresAt.UTC().Format(time.RFC3339Nano), session.LastUsedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("create authentication session: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, created_at, expires_at, revoked_at, last_used_at
		FROM auth_sessions
		WHERE token_hash = ?
	`, tokenHash)
	return scanSession(row)
}

func (r *SQLiteRepository) TouchSession(ctx context.Context, id string, lastUsedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE auth_sessions SET last_used_at = ? WHERE id = ?`, lastUsedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("touch authentication session: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) RevokeSession(ctx context.Context, id string, revokedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at = ? WHERE id = ?`, revokedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("revoke authentication session: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) initialize() error {
	_, err := r.db.Exec(`PRAGMA foreign_keys = ON`)
	if err != nil {
		return fmt.Errorf("enable authentication foreign keys: %w", err)
	}
	_, err = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("initialize users: %w", err)
	}
	_, err = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS auth_sessions (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			revoked_at TEXT,
			last_used_at TEXT NOT NULL,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		return fmt.Errorf("initialize authentication sessions: %w", err)
	}
	_, err = r.db.Exec(`CREATE INDEX IF NOT EXISTS auth_sessions_token_hash ON auth_sessions(token_hash)`)
	if err != nil {
		return fmt.Errorf("initialize authentication token index: %w", err)
	}

	return nil
}

func scanSession(row *sql.Row) (Session, error) {
	var session Session
	var createdAt string
	var expiresAt string
	var revokedAt sql.NullString
	var lastUsedAt string
	if err := row.Scan(&session.ID, &session.UserID, &session.TokenHash, &createdAt, &expiresAt, &revokedAt, &lastUsedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return Session{}, fmt.Errorf("get authentication session: %w", err)
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return Session{}, fmt.Errorf("parse authentication session creation time: %w", err)
	}
	parsedExpiresAt, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return Session{}, fmt.Errorf("parse authentication session expiry time: %w", err)
	}
	parsedLastUsedAt, err := time.Parse(time.RFC3339Nano, lastUsedAt)
	if err != nil {
		return Session{}, fmt.Errorf("parse authentication session last used time: %w", err)
	}
	session.CreatedAt = parsedCreatedAt
	session.ExpiresAt = parsedExpiresAt
	session.LastUsedAt = parsedLastUsedAt
	if revokedAt.Valid {
		parsedRevokedAt, err := time.Parse(time.RFC3339Nano, revokedAt.String)
		if err != nil {
			return Session{}, fmt.Errorf("parse authentication session revocation time: %w", err)
		}
		session.RevokedAt = &parsedRevokedAt
	}

	return session, nil
}

func sqliteConstraint(err error) bool {
	return strings.Contains(err.Error(), "constraint failed")
}
