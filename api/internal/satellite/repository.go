package satellite

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, s *Satellite) (*Satellite, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Satellite, error)
	List(ctx context.Context) ([]*Satellite, error)
	Update(ctx context.Context, id uuid.UUID, region string) (*Satellite, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, s *Satellite) (*Satellite, error) {
	query := `
		INSERT INTO satellites (name, region, status, managed_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, region, status, managed_by, last_seen_at, created_at
	`
	var out Satellite
	err := r.db.QueryRow(ctx, query, s.Name, s.Region, s.Status, s.ManagedBy).Scan(
		&out.ID, &out.Name, &out.Region, &out.Status, &out.ManagedBy, &out.LastSeenAt, &out.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*Satellite, error) {
	query := `
		SELECT id, name, region, status, managed_by, last_seen_at, created_at
		FROM satellites WHERE id = $1
	`
	var out Satellite
	err := r.db.QueryRow(ctx, query, id).Scan(
		&out.ID, &out.Name, &out.Region, &out.Status, &out.ManagedBy, &out.LastSeenAt, &out.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]*Satellite, error) {
	query := `
		SELECT id, name, region, status, managed_by, last_seen_at, created_at
		FROM satellites ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Satellite
	for rows.Next() {
		var s Satellite
		if err := rows.Scan(&s.ID, &s.Name, &s.Region, &s.Status, &s.ManagedBy, &s.LastSeenAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &s)
	}
	return result, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, id uuid.UUID, region string) (*Satellite, error) {
	query := `
		UPDATE satellites SET region = $1
		WHERE id = $2
		RETURNING id, name, region, status, managed_by, last_seen_at, created_at
	`
	var out Satellite
	err := r.db.QueryRow(ctx, query, region, id).Scan(
		&out.ID, &out.Name, &out.Region, &out.Status, &out.ManagedBy, &out.LastSeenAt, &out.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &out, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM satellites WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
