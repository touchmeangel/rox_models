package run

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxIDAttempts     = 3
	pgUniqueViolation = "23505"
)

type RunStore struct {
	pool *pgxpool.Pool
}

func NewRunStore(pool *pgxpool.Pool) *RunStore {
	return &RunStore{pool: pool}
}

func (s *RunStore) GetRun(ctx context.Context, runID string) (Run, error) {
	const query = `SELECT id, user_id, status, workspace_folder FROM users WHERE id = $1`

	rows, err := s.pool.Query(ctx, query, runID)
	if err != nil {
		return Run{}, fmt.Errorf("querying run %s: %w", runID, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[runRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, fmt.Errorf("user %s: %w", runID, ErrNotFound)
		}
		return Run{}, fmt.Errorf("scanning run %s: %w", runID, err)
	}

	return row.toSDK(), nil
}

type Page struct {
	Runs       []Run
	NextCursor string
}

type cursor struct {
	CreatedAt time.Time `json:"created_at"`
}

func encodeCursor(c cursor) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (cursor, error) {
	var c cursor
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return cursor{}, fmt.Errorf("invalid cursor: %w", err)
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return cursor{}, fmt.Errorf("invalid cursor: %w", err)
	}
	return c, nil
}

func (s *RunStore) ListByUser(ctx context.Context, userID, cursorStr string, limit int) (Page, error) {
	var (
		query string
		args  []any
	)
	if cursorStr == "" {
		query = `
			SELECT id, user_id, status, workspace_folder, created_at
			FROM runs
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []any{userID, limit + 1} // fetch one extra to know if there's a next page
	} else {
		c, err := decodeCursor(cursorStr)
		if err != nil {
			return Page{}, err
		}
		query = `
			SELECT id, user_id, status, workspace_folder, created_at
			FROM runs
			WHERE user_id = $1 AND (created_at) < ($2)
			ORDER BY created_at DESC
			LIMIT $4
		`
		args = []any{userID, c.CreatedAt, limit + 1}
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return Page{}, fmt.Errorf("querying runs for user %s: %w", userID, err)
	}

	runRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[runRow])
	if err != nil {
		return Page{}, fmt.Errorf("scanning runs for user %s: %w", userID, err)
	}

	hasMore := len(runRows) > limit
	if hasMore {
		runRows = runRows[:limit]
	}

	runs := make([]Run, len(runRows))
	for i, r := range runRows {
		runs[i] = r.toSDK()
	}

	var next string
	if hasMore {
		last := runRows[len(runRows)-1]
		next = encodeCursor(cursor{CreatedAt: last.CreatedAt})
	}

	return Page{Runs: runs, NextCursor: next}, nil
}
