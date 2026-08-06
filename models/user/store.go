package user

import "context"

type Store interface {
	GetByEmail(ctx context.Context, email string) (User, error)
}
