package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, config Config) (*DB, error) {
	connString := config.ConnectionString()
	pgPool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database, Error: %w", err)
	}

	return &DB{
		Pool: pgPool,
	}, nil
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

func (db *DB) Ping(ctx context.Context) error {
	if err := db.Pool.Ping(ctx); err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
