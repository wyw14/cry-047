package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct{ pool *pgxpool.Pool }

func Open(ctx context.Context, url string) (*Database, error) {
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	config.MaxConns = 12
	config.MinConns = 2
	config.MaxConnIdleTime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}
	database := &Database{pool: pool}
	if err := database.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return database, nil
}

func (d *Database) Ping(ctx context.Context) error {
	if err := d.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}

func (d *Database) Close() { d.pool.Close() }
