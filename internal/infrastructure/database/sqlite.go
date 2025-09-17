package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// SQLiteDB wraps GORM database connection
type SQLiteDB struct {
	DB *gorm.DB
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Configure GORM with SQLite driver
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Log SQL queries
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	sqliteDB := &SQLiteDB{
		DB: db,
	}

	// Run auto-migration to create tables
	if err := sqliteDB.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return sqliteDB, nil
}

// Migrate runs auto-migration for all entities
func (s *SQLiteDB) Migrate() error {
	return s.DB.AutoMigrate(
		&entities.Transaction{},
		&entities.ExchangeRate{},
	)
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDB returns the underlying GORM database instance
func (s *SQLiteDB) GetDB() *gorm.DB {
	return s.DB
}