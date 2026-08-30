package coordinator

import (
	"time"
)

type Coordinator struct {
	ID        string
	RunID     string
	Active    bool
	Completed bool
	Error     string
	Retriable bool
	CreatedAt time.Time
}

type coordinatorRaw struct {
	ID        string    `db:"id"`
	RunID     string    `db:"run_id"`
	Active    bool      `db:"active"`
	Completed bool      `db:"completed"`
	Error     string    `db:"error"`
	Retriable bool      `db:"retriable"`
	CreatedAt time.Time `db:"created_at"`
}

func (w coordinatorRaw) toSDK() Coordinator {
	return Coordinator(w)
}
