// Package postgres implements the users repository on PostgreSQL (pgx).
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/app/users/internal/domain"
)

// invalidTextRepresentation is raised by Postgres for malformed input such as a bad UUID.
const invalidTextRepresentation = "22P02"

const userColumns = `id::text, telegram_id, username, first_name, last_name, phone, created_at, updated_at`

// UserRepository stores users in the "users" table.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// List implements usecase.UserRepository.
func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+userColumns+` FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	users, err := pgx.CollectRows(rows, scanUser)
	if err != nil {
		return nil, fmt.Errorf("scan users: %w", err)
	}
	return users, nil
}

// GetByID implements usecase.UserRepository.
func (r *UserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("query user: %w", err)
	}
	user, err := pgx.CollectExactlyOneRow(rows, scanUser)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || isInvalidInput(err) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}
	return user, nil
}

// UpsertByTelegramID implements usecase.UserRepository.
func (r *UserRepository) UpsertByTelegramID(
	ctx context.Context,
	profile domain.TelegramProfile,
) (domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		INSERT INTO users (telegram_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (telegram_id) DO UPDATE SET
			username   = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name  = EXCLUDED.last_name,
			updated_at = now()
		RETURNING `+userColumns,
		profile.TelegramID, profile.Username, profile.FirstName, profile.LastName,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}
	user, err := pgx.CollectExactlyOneRow(rows, scanUser)
	if err != nil {
		return domain.User{}, fmt.Errorf("scan upserted user: %w", err)
	}
	return user, nil
}

func scanUser(row pgx.CollectableRow) (domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.Phone,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func isInvalidInput(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == invalidTextRepresentation
}
