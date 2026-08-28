package run

import (
	"errors"
	"time"
)

var ErrInvalidCursor = errors.New("invalid cursor")

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}
