package worker

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

type WorkerStore struct {
	pool *pgxpool.Pool
}

func NewWorkerStore(pool *pgxpool.Pool) *WorkerStore {
	return &WorkerStore{pool: pool}
}

func (s *WorkerStore) GetByID(ctx context.Context, workerID string) (Worker, error) {
	const query = `
		SELECT id, run_id, mission_id, active, completed, error, retriable, created_at
		FROM workers WHERE id = $1
	`

	rows, err := s.pool.Query(ctx, query, workerID)
	if err != nil {
		return Worker{}, fmt.Errorf("querying worker %s: %w", workerID, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[workerRaw])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Worker{}, fmt.Errorf("worker %s: %w", workerID, ErrNotFound)
		}
		return Worker{}, fmt.Errorf("scanning worker %s: %w", workerID, err)
	}

	return row.toSDK(), nil
}

func (s *WorkerStore) CreateWorker(ctx context.Context, runID, missionID string) (Worker, error) {
	const insertQuery = `
		INSERT INTO workers (id, run_id, mission_id, active, completed, error, retriable, created_at)
		VALUES ($1, $2, $3, false, false, '', false, $4)
		RETURNING id, run_id, mission_id, active, completed, error, retriable, created_at
	`

	createdAt := time.Now().UTC()

	for attempt := 1; ; attempt++ {
		id, err := idgen.New("worker")
		if err != nil {
			return Worker{}, fmt.Errorf("create worker: generate id: %w", err)
		}

		rows, err := s.pool.Query(ctx, insertQuery, id, runID, missionID, createdAt)
		if err == nil {
			row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[workerRaw])
			if err != nil {
				return Worker{}, fmt.Errorf("create worker: scan inserted worker: %w", err)
			}
			return row.toSDK(), nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == idgen.PGUniqueViolation && pgErr.ConstraintName == "workers_pkey" {
			if attempt < idgen.MaxIDAttempts {
				continue
			}
			return Worker{}, fmt.Errorf("create worker: id generator collided %d times in a row (check entropy source): %w", attempt, err)
		}
		return Worker{}, fmt.Errorf("create worker: insert worker: %w", err)
	}
}

func (s *WorkerStore) UpdateActive(ctx context.Context, workerID string, active bool) (bool, error) {
	const query = `UPDATE workers SET active = $1 WHERE id = $2`
	tag, err := s.pool.Exec(ctx, query, active, workerID)
	if err != nil {
		return false, fmt.Errorf("update active for worker %s: %w", workerID, err)
	}
	return tag.RowsAffected() == 1, nil
}

func (s *WorkerStore) UpdateCompleted(ctx context.Context, workerID string, errMsg string, retriable bool) (bool, error) {
	const query = `
		UPDATE workers
		SET completed = true, error = $1, retriable = $2
		WHERE id = $3
	`
	tag, err := s.pool.Exec(ctx, query, errMsg, retriable, workerID)
	if err != nil {
		return false, fmt.Errorf("update completed for worker %s: %w", workerID, err)
	}
	return tag.RowsAffected() == 1, nil
}
