package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

type userRow struct {
	ID       string   `db:"id"`
	Email    string   `db:"email"`
	Username string   `db:"username"`
	Roles    []string `db:"roles"`
}

func (r userRow) toSDK() User {
	roles := make([]Role, len(r.Roles))
	for i, ro := range r.Roles {
		roles[i] = Role(ro)
	}
	return User{
		ID:       r.ID,
		Email:    r.Email,
		Username: r.Username,
		Roles:    roles,
	}
}

func (s *UserStore) GetProfile(ctx context.Context, userID string) (User, error) {
	const query = `SELECT id, email, username, roles FROM users WHERE id = $1`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return User{}, fmt.Errorf("querying user profile %s: %w", userID, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user %s: %w", userID, ErrNotFound)
		}
		return User{}, fmt.Errorf("scanning user profile %s: %w", userID, err)
	}

	return row.toSDK(), nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (User, error) {
	const query = `SELECT id, email, username, roles FROM users WHERE email = $1`

	rows, err := s.pool.Query(ctx, query, email)
	if err != nil {
		return User{}, fmt.Errorf("querying user by email %s: %w", email, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user %s: %w", email, ErrNotFound)
		}
		return User{}, fmt.Errorf("scanning user profile %s: %w", email, err)
	}

	return row.toSDK(), nil
}
