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
	ID    string
	Email string
	Roles []Role
}
