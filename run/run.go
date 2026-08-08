package run

import (
	"errors"
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
}
