package run

import (
	"time"
)

type Status string

const (
	StatusPendingUpload      Status = "pending_upload"
	StatusUploaded           Status = "uploaded"
	StatusStarting           Status = "starting"
	StatusRunningCoordinator Status = "running_coordinator"
	StatusRunningWorkers     Status = "running_workers"
	StatusPaused             Status = "paused"
	StatusComplete           Status = "complete"
)

type Run struct {
	ID            string
	Name          string
	UserID        string
	Status        Status
	WorkspaceName string
	CreatedAt     time.Time
}

type runRow struct {
	ID                   string    `db:"id"`
	Name                 string    `db:"name"`
	UserID               string    `db:"user_id"`
	Status               string    `db:"status"`
	WorkerCount          int       `db:"worker_count"`
	CompletedWorkerCount int       `db:"completed_worker_count"`
	ActiveWorkerCount    int       `db:"active_worker_count"`
	WorkspaceName        string    `db:"workspace_name"`
	CreatedAt            time.Time `db:"created_at"`
}

func (r runRow) toSDK() Run {
	return Run{
		ID:            r.ID,
		Name:          r.Name,
		UserID:        r.UserID,
		Status:        Status(r.Status),
		WorkspaceName: r.WorkspaceName,
		CreatedAt:     r.CreatedAt,
	}
}
