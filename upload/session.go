package upload

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Offset         int64
	CompletedParts []CompletedPart
}

type sessionRaw struct {
	RunID          string `db:"run_id"`
	UserID         string `db:"user_id"`
	UploadID       string `db:"upload_id"`
	Offset         int64  `db:"offset_bytes"`
	CompletedParts []byte `db:"completed_parts"`
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
	}, nil
}
