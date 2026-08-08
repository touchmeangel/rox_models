package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/touchmeangel/rox_models_go/idgen"
	"github.com/touchmeangel/rox_models_go/signup"
)

const (
	maxIDAttempts     = 3
	pgUniqueViolation = "23505"
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
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

func (s *UserStore) GetByUsername(ctx context.Context, username string) (User, error) {
	const query = `SELECT id, email, username, roles FROM users WHERE username = $1`

	rows, err := s.pool.Query(ctx, query, username)
	if err != nil {
		return User{}, fmt.Errorf("querying user by username %s: %w", username, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user %s: %w", username, ErrNotFound)
		}
		return User{}, fmt.Errorf("scanning user profile %s: %w", username, err)
	}

	return row.toSDK(), nil
}

func (s *UserStore) GetCredentialsByEmail(ctx context.Context, email string) (Credentials, error) {
	const query = `SELECT id, email, username, roles, password_hash FROM users WHERE email = $1`

	rows, err := s.pool.Query(ctx, query, email)
	if err != nil {
		return Credentials{}, fmt.Errorf("querying credentials by email %s: %w", email, err)
	}

	row, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[credentialsRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Credentials{}, fmt.Errorf("user %s: %w", email, ErrNotFound)
		}
		return Credentials{}, fmt.Errorf("scanning credentials %s: %w", email, err)
	}

	return row.toSDK(), nil
}

func (s *UserStore) SignupStatus(ctx context.Context) (signup.SignupMode, error) {
	const query = `SELECT admin_claimed, open_signup_enabled FROM system_settings`

	var adminClaimed, openSignup bool
	err := s.pool.QueryRow(ctx, query).Scan(&adminClaimed, &openSignup)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("signup status: system_settings row missing — check migrations ran")
		}
		return "", fmt.Errorf("signup status: %w", err)
	}

	switch {
	case !adminClaimed:
		return signup.SignupModeBootstrap, nil
	case openSignup:
		return signup.SignupModeOpen, nil
	default:
		return signup.SignupModeInviteOnly, nil
	}
}

var (
	ErrSignupClosed  = errors.New("signup is currently closed")
	ErrEmailTaken    = errors.New("email is already registered")
	ErrUsernameTaken = errors.New("username is already registered")
)

func (s *UserStore) Register(ctx context.Context, email, username, passwordHash string) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("register: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `UPDATE system_settings SET admin_claimed = true WHERE admin_claimed = false`)
	if err != nil {
		return User{}, fmt.Errorf("register: claim admin slot: %w", err)
	}
	becameAdmin := tag.RowsAffected() == 1

	role := RoleUser
	if becameAdmin {
		role = RoleAdmin
	} else {
		var openSignup bool
		err := tx.QueryRow(ctx, `SELECT open_signup_enabled FROM system_settings`).Scan(&openSignup)
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("register: system_settings row missing — check migrations ran")
		}
		if err != nil {
			return User{}, fmt.Errorf("register: check signup mode: %w", err)
		}
		if !openSignup {
			return User{}, ErrSignupClosed
		}
	}

	const insertQuery = `
		INSERT INTO users (id, email, email_verified, username, password_hash, roles)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, username, roles
	`

	var row userRow
	for attempt := 1; ; attempt++ {
		id, err := idgen.New("usr")
		if err != nil {
			return User{}, fmt.Errorf("register: generate id: %w", err)
		}

		rows, err := tx.Query(ctx, insertQuery, id, email, false, username, passwordHash, []string{string(role)})
		if err == nil {
			row, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[userRow])
			if err != nil {
				return User{}, fmt.Errorf("register: scan inserted user: %w", err)
			}
			break
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return User{}, ErrEmailTaken
			case "users_username_key":
				return User{}, ErrUsernameTaken
			case "users_pkey":
				if attempt < maxIDAttempts {
					continue
				}
				return User{}, fmt.Errorf("register: id generator collided %d times in a row (check entropy source): %w", attempt, err)
			}
		}
		return User{}, fmt.Errorf("register: insert user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("register: commit: %w", err)
	}

	return row.toSDK(), nil
}

func (s *UserStore) SetQuota(ctx context.Context, userID string, quotaGiB int) error {
	const query = `UPDATE users SET quota_gib = $1 WHERE id = $2`

	cmdTag, err := s.pool.Exec(ctx, query, quotaGiB, userID)
	if err != nil {
		return fmt.Errorf("updating quota for user %s: %w", userID, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user %s: %w", userID, ErrNotFound)
	}

	return nil
}
