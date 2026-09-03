package upload

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound       = errors.New("upload session not found")
	ErrOffsetMismatch = errors.New("offset mismatch")
)

type CompletedPart struct {
	PartNumber int32  `json:"part_number"`
	ETag       string `json:"etag"`
}

type Session struct {
	RunID          string
	UserID         string
	UploadID       string
	Offset         uint64
	CompletedParts []CompletedPart
	ReserveAmount  uint64
	LastActiveAt   time.Time
}

type sessionRaw struct {
	RunID          string    `db:"run_id"`
	UserID         string    `db:"user_id"`
	UploadID       string    `db:"upload_id"`
	Offset         uint64    `db:"offset_bytes"`
	CompletedParts []byte    `db:"completed_parts"`
	ReserveAmount  uint64    `db:"reserve_amount"`
	LastActiveAt   time.Time `db:"last_active_at"`
}

func (s sessionRaw) toSDK() (Session, error) {
	var completedParts []CompletedPart

	if err := json.Unmarshal(s.CompletedParts, &completedParts); err != nil {
		return Session{}, fmt.Errorf("decode completed parts: %w", err)
	}

	return Session{
		RunID:          s.RunID,
		UserID:         s.UserID,
		UploadID:       s.UploadID,
		Offset:         s.Offset,
		CompletedParts: completedParts,
		ReserveAmount:  s.ReserveAmount,
		LastActiveAt:   s.LastActiveAt,
	}, nil
}
