package user

import (
	"errors"
)

var ErrNotFound = errors.New("user not found")

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
}

type userRow struct {
	ID            string   `db:"id"`
	Email         string   `db:"email"`
	EmailVerified bool     `db:"email_verified"`
	Username      string   `db:"username"`
	Roles         []string `db:"roles"`
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
	}
}
