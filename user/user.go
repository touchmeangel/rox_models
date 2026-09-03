package user

import (
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID            string
	Email         string
	EmailVerified bool
	Username      string
	Roles         []Role
	QuotaBytes    uint64
	UsedBytes     uint64
	RegisteredAt  time.Time
}

type userRow struct {
	ID            string    `db:"id"`
	Email         string    `db:"email"`
	EmailVerified bool      `db:"email_verified"`
	Username      string    `db:"username"`
	Roles         []string  `db:"roles"`
	QuotaBytes    uint64    `db:"quota_bytes"`
	UsedBytes     uint64    `db:"used_bytes"`
	RegisteredAt  time.Time `db:"registered_at"`
}

func (r userRow) toSDK() User {
	roles := make([]Role, len(r.Roles))
	for i, ro := range r.Roles {
		roles[i] = Role(ro)
	}
	return User{
		ID:            r.ID,
		Email:         r.Email,
		EmailVerified: r.EmailVerified,
		Username:      r.Username,
		Roles:         roles,
		QuotaBytes:    r.QuotaBytes,
		UsedBytes:     r.UsedBytes,
		RegisteredAt:  r.RegisteredAt,
	}
}
