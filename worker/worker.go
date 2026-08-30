package worker

import (
	"time"
)

type Worker struct {
	ID        string
	RunID     string
	MissionID string
	Active    bool
	Completed bool
	Error     string
	Retriable bool
	CreatedAt time.Time
}

type workerRaw struct {
	ID        string    `db:"id"`
	RunID     string    `db:"run_id"`
	MissionID string    `db:"mission_id"`
	Active    bool      `db:"active"`
	Completed bool      `db:"completed"`
	Error     string    `db:"error"`
	Retriable bool      `db:"retriable"`
	CreatedAt time.Time `db:"created_at"`
}

func (w workerRaw) toSDK() Worker {
	return Worker(w)
}
