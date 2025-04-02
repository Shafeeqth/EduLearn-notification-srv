package database

import (
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

type DB struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewDB(dsn string, logger *zap.Logger) (*DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 zapgorm2.New(logger),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to database")
	return &DB{db: db, logger: logger}, nil
}

func (d *DB) DB() *gorm.DB {
	return d.db
}

func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		d.logger.Error("Failed to get SQL DB", zap.Error(err))
		return err
	}
	if err := sqlDB.Close(); err != nil {
		d.logger.Error("Failed to close database", zap.Error(err))
		return err
	}
	d.logger.Info("Database connection closed")
	return nil
}
