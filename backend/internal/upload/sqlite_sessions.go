package upload

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var ErrMissingDatabasePath = errors.New("missing upload database path")
var ErrInvalidSessionRecord = errors.New("invalid upload session record")

type SQLiteSessionRepository struct {
	db *sql.DB
}

func NewSQLiteSessionRepository(databasePath string) (*SQLiteSessionRepository, error) {
	if databasePath == "" {
		return nil, ErrMissingDatabasePath
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open upload database: %w", err)
	}

	repository := &SQLiteSessionRepository{db: db}
	if err := repository.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return repository, nil
}

func (r *SQLiteSessionRepository) Close() error {
	return r.db.Close()
}

func (r *SQLiteSessionRepository) Create(ctx context.Context, session UploadSession) error {
	if err := validateSession(session); err != nil {
		return err
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO upload_sessions (
			id, owner_id, display_name, blob_name, declared_size, extension,
			content_type, idempotency_key, created_at, expires_at, status, version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.OwnerID, session.DisplayName, session.BlobName,
		session.DeclaredSize, session.Extension, session.ContentType,
		session.IdempotencyKey, session.CreatedAt.UTC().Format(time.RFC3339Nano),
		session.ExpiresAt.UTC().Format(time.RFC3339Nano),
		session.Status, session.Version)
	if err != nil {
		return fmt.Errorf("create upload session: %w", err)
	}

	return nil
}

func (r *SQLiteSessionRepository) GetByID(ctx context.Context, id string) (UploadSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, display_name, blob_name, declared_size, extension,
			content_type, idempotency_key, created_at, expires_at, status, version
		FROM upload_sessions
		WHERE id = ?
	`, id)

	return scanSession(row)
}

func (r *SQLiteSessionRepository) GetByIdempotencyKey(ctx context.Context, ownerID string, key string) (UploadSession, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_id, display_name, blob_name, declared_size, extension,
			content_type, idempotency_key, created_at, expires_at, status, version
		FROM upload_sessions
		WHERE owner_id = ? AND idempotency_key = ?
	`, ownerID, key)

	return scanSession(row)
}

func scanSession(row *sql.Row) (UploadSession, error) {
	var session UploadSession
	var createdAt string
	var expiresAt string
	if err := row.Scan(
		&session.ID,
		&session.OwnerID,
		&session.DisplayName,
		&session.BlobName,
		&session.DeclaredSize,
		&session.Extension,
		&session.ContentType,
		&session.IdempotencyKey,
		&createdAt,
		&expiresAt,
		&session.Status,
		&session.Version,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return UploadSession{}, ErrSessionNotFound
		}
		return UploadSession{}, fmt.Errorf("get upload session: %w", err)
	}

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return UploadSession{}, fmt.Errorf("%w: %w", ErrInvalidSessionRecord, err)
	}
	parsedExpiresAt, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return UploadSession{}, fmt.Errorf("%w: %w", ErrInvalidSessionRecord, err)
	}
	session.CreatedAt = parsedCreatedAt
	session.ExpiresAt = parsedExpiresAt

	return session, nil
}

func (r *SQLiteSessionRepository) initialize() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS upload_sessions (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			display_name TEXT NOT NULL,
			blob_name TEXT NOT NULL UNIQUE,
			declared_size INTEGER NOT NULL,
			extension TEXT NOT NULL,
			content_type TEXT NOT NULL,
			idempotency_key TEXT NOT NULL,
			created_at TEXT NOT NULL,
			expires_at TEXT NOT NULL,
			status TEXT NOT NULL,
			version INTEGER NOT NULL,
			UNIQUE(owner_id, idempotency_key)
		)
	`)
	if err != nil {
		return fmt.Errorf("initialize upload sessions: %w", err)
	}
	hasIdempotencyKey, err := r.hasColumn("idempotency_key")
	if err != nil {
		return err
	}
	if !hasIdempotencyKey {
		_, err = r.db.Exec(`ALTER TABLE upload_sessions ADD COLUMN idempotency_key TEXT NOT NULL DEFAULT ''`)
		if err != nil {
			return fmt.Errorf("add upload session idempotency key: %w", err)
		}
		_, err = r.db.Exec(`UPDATE upload_sessions SET idempotency_key = id WHERE idempotency_key = ''`)
		if err != nil {
			return fmt.Errorf("migrate upload session idempotency keys: %w", err)
		}
	}
	_, err = r.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS upload_sessions_owner_idempotency_key ON upload_sessions(owner_id, idempotency_key)`)
	if err != nil {
		return fmt.Errorf("initialize upload session idempotency index: %w", err)
	}

	return nil
}

func (r *SQLiteSessionRepository) hasColumn(name string) (bool, error) {
	rows, err := r.db.Query(`PRAGMA table_info(upload_sessions)`)
	if err != nil {
		return false, fmt.Errorf("inspect upload session schema: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var index int
		var columnName string
		var columnType string
		var notNull int
		var defaultValue any
		var primaryKey int
		if err := rows.Scan(&index, &columnName, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, fmt.Errorf("read upload session schema: %w", err)
		}
		if columnName == name {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate upload session schema: %w", err)
	}

	return false, nil
}

func validateSession(session UploadSession) error {
	if session.ID == "" || session.OwnerID == "" || session.DisplayName == "" ||
		session.BlobName == "" || session.DeclaredSize <= 0 || session.Extension == "" ||
		session.ContentType == "" || session.IdempotencyKey == "" || session.CreatedAt.IsZero() || session.ExpiresAt.IsZero() ||
		session.Status == "" || session.Version <= 0 {
		return ErrInvalidSession
	}

	return nil
}
