package upload

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionStore struct {
	pool *pgxpool.Pool
}

func NewSessionStore(pool *pgxpool.Pool) *SessionStore {
	return &SessionStore{pool: pool}
}

func (s *SessionStore) Create(ctx context.Context, userID, runID, uploadID string) error {
	const query = `INSERT INTO upload_sessions (run_id, user_id, upload_id) VALUES ($1, $2, $3)`
	if _, err := s.pool.Exec(ctx, query, runID, userID, uploadID); err != nil {
		return fmt.Errorf("create upload session %s: %w", runID, err)
	}
	return nil
}

func (s *SessionStore) Get(ctx context.Context, runID string) (Session, error) {
	const query = `
		SELECT run_id, user_id, upload_id, offset_bytes, completed_parts
		FROM upload_sessions
		WHERE run_id = $1
	`

	rows, err := s.pool.Query(ctx, query, runID)
	if err != nil {
		return Session{}, fmt.Errorf("querying upload session %s: %w", runID, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[sessionRaw])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, fmt.Errorf("upload session %s: %w", runID, ErrNotFound)
		}

		return Session{}, fmt.Errorf("scanning upload session %s: %w", runID, err)
	}

	session, err := row.toSDK()
	if err != nil {
		return Session{}, fmt.Errorf("parsing upload session %s: %w", runID, err)
	}

	return session, nil
}

func (s *SessionStore) AppendPart(ctx context.Context, runID, userID string, expectedOffset int64, part CompletedPart, chunkLen int64) (newOffset int64, err error) {
	partJSON, err := json.Marshal(part)
	if err != nil {
		return 0, fmt.Errorf("encode part: %w", err)
	}

	const query = `
		UPDATE upload_sessions
		SET offset_bytes = offset_bytes + $1,
		    completed_parts = completed_parts || $2::jsonb,
		    last_active_at = now()
		WHERE run_id = $3 AND user_id = $4 AND offset_bytes = $5
		RETURNING offset_bytes
	`
	err = s.pool.QueryRow(ctx, query, chunkLen, partJSON, runID, userID, expectedOffset).Scan(&newOffset)
	if errors.Is(err, pgx.ErrNoRows) {
		current, getErr := s.Get(ctx, runID)
		if getErr != nil {
			return 0, ErrNotFound
		}
		return current.Offset, ErrOffsetMismatch
	}
	if err != nil {
		return 0, fmt.Errorf("append part to session %s: %w", runID, err)
	}
	return newOffset, nil
}

func (s *SessionStore) Delete(ctx context.Context, runID string) error {
	if _, err := s.pool.Exec(ctx, `DELETE FROM upload_sessions WHERE run_id = $1`, runID); err != nil {
		return fmt.Errorf("delete upload session %s: %w", runID, err)
	}
	return nil
}
