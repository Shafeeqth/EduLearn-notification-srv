package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDB(dsn string, logger *zap.Logger) (*DB, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to database")
	return &DB{pool: pool, logger: logger}, nil
}

func (db *DB) Close() {
	db.pool.Close()
	db.logger.Info("Database connection closed")
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}
