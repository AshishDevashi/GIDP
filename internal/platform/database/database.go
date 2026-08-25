package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pooled connection to Postgres used by sqlc-generated queries.
func Connect(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
