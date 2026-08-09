package run

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("run not found")

type Status string

const (
	StatusStarting Status = "starting"
	StatusPaused   Status = "paused"
	StatusComplete Status = "complete"
)

type Run struct {
	ID              string
	UserID          string
	Status          Status
	WorkspaceFolder string
	CreatedAt       time.Time
}

type runRow struct {
	ID              string    `db:"id"`
	UserID          string    `db:"user_id"`
	Status          string    `db:"status"`
	WorkspaceFolder string    `db:"workspace_folder"`
	CreatedAt       time.Time `db:"created_at"`
}

func (r runRow) toSDK() Run {
	return Run{
		ID: r.ID, UserID: r.UserID, Status: Status(r.Status),
		WorkspaceFolder: r.WorkspaceFolder, CreatedAt: r.CreatedAt,
	}
}
