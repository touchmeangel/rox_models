package task

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("run not found")

type Type string

const (
	TypeCoordinator Type = "coordinator"
	TypeWorker      Type = "worker"
)

type Worker struct {
	ID        string
	RunID     string
	Type      Type
	Active    bool
	Completed bool
	Error     string
	Retriable bool
	CreatedAt time.Time
}

type workerRaw struct {
	ID        string    `db:"id"`
	RunID     string    `db:"run_id"`
	Type      Type      `db:"type"`
	Active    bool      `db:"active"`
	Completed bool      `db:"completed"`
	Error     string    `db:"error"`
	Retriable bool      `db:"retriable"`
	CreatedAt time.Time `db:"created_at"`
}

func (w workerRaw) toSDK() Worker {
	return Worker{
		ID:        w.ID,
		RunID:     w.RunID,
		Type:      w.Type,
		Active:    w.Active,
		Completed: w.Completed,
		Error:     w.Error,
		Retriable: w.Retriable,
		CreatedAt: w.CreatedAt,
	}
}
