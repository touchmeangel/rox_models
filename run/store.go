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

var ErrNotFound = errors.New("run not found")

type RunStore struct {
	pool *pgxpool.Pool
}

func NewRunStore(pool *pgxpool.Pool) *RunStore {
	return &RunStore{pool: pool}
}

func (s *RunStore) GetByID(ctx context.Context, runID string) (Run, error) {
	const query = `SELECT id, name, user_id, status, workspace_name, created_at FROM runs WHERE id = $1`

	rows, err := s.pool.Query(ctx, query, runID)
	if err != nil {
		return Run{}, fmt.Errorf("querying run %s: %w", runID, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[runRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, fmt.Errorf("run %s: %w", runID, ErrNotFound)
		}
		return Run{}, fmt.Errorf("scanning run %s: %w", runID, err)
	}

	return row.toSDK(), nil
}

func (s *RunStore) GetByName(ctx context.Context, userID string, name string) (Run, error) {
	const query = `SELECT id, name, user_id, status, workspace_name, created_at FROM runs WHERE user_id = $1 AND name = $2`

	rows, err := s.pool.Query(ctx, query, userID, name)
	if err != nil {
		return Run{}, fmt.Errorf("querying run %s: %w", name, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[runRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, fmt.Errorf("run %s: %w", name, ErrNotFound)
		}
		return Run{}, fmt.Errorf("scanning run %s: %w", name, err)
	}

	return row.toSDK(), nil
}

type Page struct {
	Runs       []Run
	NextCursor *Cursor
}

func (s *RunStore) ListByUser(ctx context.Context, userID, cursor *Cursor, limit int) (Page, error) {
	var (
		query string
		args  []any
	)
	if cursor == nil {
		query = `
			SELECT id, name, user_id, status, workspace_name, created_at
			FROM runs
			WHERE user_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		`
		args = []any{userID, limit + 1} // fetch one extra to know if there's a next page
	} else {
		query = `
			SELECT id, name, user_id, status, workspace_name, created_at
			FROM runs
			WHERE user_id = $1 AND (created_at, id) < ($2, $3)
			ORDER BY created_at DESC, id DESC
			LIMIT $4
		`
		args = []any{userID, cursor.CreatedAt, cursor.ID, limit + 1}
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

	var next *Cursor
	if hasMore {
		last := runRows[len(runRows)-1]
		next = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	return Page{Runs: runs, NextCursor: next}, nil
}

func (s *RunStore) CreateRun(ctx context.Context, name string, userID string) (Run, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Run{}, fmt.Errorf("create run: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	createdAt := time.Now().UTC()

	const insertQuery = `
		INSERT INTO runs (id, name, user_id, status, workspace_name, created_at)
		VALUES ($1, $2, $3, $4, '', 0, $5)
		RETURNING id, name, user_id, status, workspace_name, created_at
	`

	var row runRow
	for attempt := 1; ; attempt++ {
		id, err := idgen.New("run")
		if err != nil {
			return Run{}, fmt.Errorf("create run: generate id: %w", err)
		}

		rows, err := tx.Query(ctx, insertQuery, id, name, userID, string(StatusPendingUpload), createdAt)
		if err == nil {
			row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[runRow])
			if err != nil {
				return Run{}, fmt.Errorf("create run: scan inserted run: %w", err)
			}
			break
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == idgen.PGUniqueViolation && pgErr.ConstraintName == "runs_pkey" {
			if attempt < idgen.MaxIDAttempts {
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

func (s *RunStore) DeleteRun(ctx context.Context, runID, userID string) (bool, error) {
	const query = `DELETE FROM runs WHERE id = $1 AND user_id = $2`
	tag, err := s.pool.Exec(ctx, query, runID, userID)
	if err != nil {
		return false, fmt.Errorf("delete run %s: %w", runID, err)
	}
	return tag.RowsAffected() == 1, nil
}

func (s *RunStore) SetWorkspaceFolder(ctx context.Context, runID, folderName string) (bool, error) {
	const query = `
		UPDATE runs SET workspace_name = $1, status = $2
		WHERE id = $3 AND workspace_name = ''
	`
	tag, err := s.pool.Exec(ctx, query, folderName, string(StatusUploaded), runID)
	if err != nil {
		return false, fmt.Errorf("set workspace folder for run %s: %w", runID, err)
	}
	return tag.RowsAffected() == 1, nil
}

func (s *RunStore) UpdateStatus(ctx context.Context, runID string, to Status) (bool, error) {
	const query = `
		UPDATE runs SET status = $1
		WHERE id = $2
	`
	tag, err := s.pool.Exec(ctx, query, string(to), runID)
	if err != nil {
		return false, fmt.Errorf("update status for run %s: %w", runID, err)
	}
	return tag.RowsAffected() == 1, nil
}
