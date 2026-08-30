package coordinator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/touchmeangel/rox_models/idgen"
)

var ErrNotFound = errors.New("run not found")

type CoordinatorStore struct {
	pool *pgxpool.Pool
}

func NewCoordinatorStore(pool *pgxpool.Pool) *CoordinatorStore {
	return &CoordinatorStore{pool: pool}
}

func (s *CoordinatorStore) GetByID(ctx context.Context, coordinatorID string) (Coordinator, error) {
	const query = `
		SELECT id, run_id, active, completed, error, retriable, created_at
		FROM coordinators WHERE id = $1
	`

	rows, err := s.pool.Query(ctx, query, coordinatorID)
	if err != nil {
		return Coordinator{}, fmt.Errorf("querying coordinator %s: %w", coordinatorID, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[coordinatorRaw])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Coordinator{}, fmt.Errorf("coordinator %s: %w", coordinatorID, ErrNotFound)
		}
		return Coordinator{}, fmt.Errorf("scanning coordinator %s: %w", coordinatorID, err)
	}

	return row.toSDK(), nil
}

func (s *CoordinatorStore) CreateCoordinator(ctx context.Context, runID string) (Coordinator, error) {
	const insertQuery = `
		INSERT INTO coordinators (id, run_id, active, completed, error, retriable, created_at)
		VALUES ($1, $2, false, false, '', false, $3)
		RETURNING id, run_id, active, completed, error, retriable, created_at
	`

	createdAt := time.Now().UTC()

	for attempt := 1; ; attempt++ {
		id, err := idgen.New("coordinator")
		if err != nil {
			return Coordinator{}, fmt.Errorf("create coordinator: generate id: %w", err)
		}

		rows, err := s.pool.Query(ctx, insertQuery, id, runID, createdAt)
		if err == nil {
			row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[coordinatorRaw])
			if err != nil {
				return Coordinator{}, fmt.Errorf("create coordinator: scan inserted coordinator: %w", err)
			}
			return row.toSDK(), nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == idgen.PGUniqueViolation && pgErr.ConstraintName == "coordinators_pkey" {
			if attempt < idgen.MaxIDAttempts {
				continue
			}
			return Coordinator{}, fmt.Errorf("create coordinator: id generator collided %d times in a row (check entropy source): %w", attempt, err)
		}
		return Coordinator{}, fmt.Errorf("create coordinator: insert coordinator: %w", err)
	}
}

func (s *CoordinatorStore) UpdateActive(ctx context.Context, coordinatorID string, active bool) (bool, error) {
	const query = `UPDATE coordinators SET active = $1 WHERE id = $2`
	tag, err := s.pool.Exec(ctx, query, active, coordinatorID)
	if err != nil {
		return false, fmt.Errorf("update active for coordinator %s: %w", coordinatorID, err)
	}
	return tag.RowsAffected() == 1, nil
}

func (s *CoordinatorStore) UpdateCompleted(ctx context.Context, coordinatorID string, errMsg string, retriable bool) (bool, error) {
	const query = `
		UPDATE coordinators
		SET completed = true, error = $1, retriable = $2
		WHERE id = $3
	`
	tag, err := s.pool.Exec(ctx, query, errMsg, retriable, coordinatorID)
	if err != nil {
		return false, fmt.Errorf("update completed for coordinator %s: %w", coordinatorID, err)
	}
	return tag.RowsAffected() == 1, nil
}
