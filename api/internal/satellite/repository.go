package satellite

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{DB: db}
}

func (r *PostgresRepository) Create(ctx context.Context, s *Satellite) (*Satellite, error) {
	query := `
		INSERT INTO satellites (name, region, status, managed_by)
		VALUES ($1, $2, 'Pending', $3)
		RETURNING id, name, region, status, managed_by, last_seen_at, created_at
	`
	var out Satellite
	err := r.DB.QueryRow(ctx, query, s.Name, s.Region, s.ManagedBy).Scan(
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
	err := r.DB.QueryRow(ctx, query, id).Scan(
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
	rows, err := r.DB.Query(ctx, query)
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
