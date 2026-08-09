package run

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/touchmeangel/rox_models_go/idgen"
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
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		args = []any{userID, limit + 1} // fetch one extra to know if there's a next page
	} else {
		c, err := DecodeCursor(cursorStr)
		if err != nil {
			return Page{}, err
		}
		query = `
			SELECT id, user_id, status, workspace_folder, created_at
			FROM runs
			WHERE user_id = $1 AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`
		args = []any{userID, c.CreatedAt, c.ID, limit + 1}
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
		next = EncodeCursor(Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
	}

	return Page{Runs: runs, NextCursor: next}, nil
}

func (s *RunStore) CreateRun(ctx context.Context, userID, folderName string) (Run, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Run{}, fmt.Errorf("create run: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	createdAt := time.Now().UTC()

	const insertQuery = `
		INSERT INTO runs (id, user_id, status, workspace_folder, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, status, workspace_folder, created_at
	`

	var row runRow
	for attempt := 1; ; attempt++ {
		id, err := idgen.New("run")
		if err != nil {
			return Run{}, fmt.Errorf("create run: generate id: %w", err)
		}

		rows, err := tx.Query(ctx, insertQuery, id, userID, string(StatusStarting), folderName, createdAt)
		if err == nil {
			row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[runRow])
			if err != nil {
				return Run{}, fmt.Errorf("create run: scan inserted run: %w", err)
			}
			break
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation && pgErr.ConstraintName == "runs_pkey" {
			if attempt < maxIDAttempts {
				continue
			}
			return Run{}, fmt.Errorf("create run: id generator collided %d times in a row (check entropy source): %w", attempt, err)
		}
		return Run{}, fmt.Errorf("create run: insert run: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Run{}, fmt.Errorf("create run: commit: %w", err)
	}

	return row.toSDK(), nil
}
