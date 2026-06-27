package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, u *User) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, u *User) (*User, error) {
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, username, password_hash, role, created_at
	`
	var out User
	err := r.db.QueryRow(ctx, query, u.Username, u.PasswordHash, u.Role).Scan(
		&out.ID, &out.Username, &out.PasswordHash, &out.Role, &out.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PostgresRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE username = $1`
	var out User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&out.ID, &out.Username, &out.PasswordHash, &out.Role, &out.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users WHERE id = $1`
	var out User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&out.ID, &out.Username, &out.PasswordHash, &out.Role, &out.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]*User, error) {
	query := `SELECT id, username, password_hash, role, created_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &u)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
