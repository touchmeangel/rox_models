package idgen

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/btcsuite/btcutil/base58"
)

func New(prefix string) (string, error) {
	var buf [22]byte

	ms := time.Now().UnixMilli()
	buf[0] = byte(ms >> 40)
	buf[1] = byte(ms >> 32)
	buf[2] = byte(ms >> 24)
	buf[3] = byte(ms >> 16)
	buf[4] = byte(ms >> 8)
	buf[5] = byte(ms)

	if _, err := rand.Read(buf[6:]); err != nil {
		return "", fmt.Errorf("idgen: read random bytes: %w", err)
	}

	return prefix + "_" + base58.Encode(buf[:]), nil
}
