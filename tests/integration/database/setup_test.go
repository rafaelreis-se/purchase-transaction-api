package database_test

import (
	"testing"

	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/database"
	"github.com/stretchr/testify/require"
)

// setupInMemoryTestDB creates an in-memory SQLite database for faster tests
func setupInMemoryTestDB(t *testing.T) (*database.SQLiteDB, func()) {
	// Use in-memory SQLite database (faster for tests)
	db, err := database.NewSQLiteDB(":memory:")
	require.NoError(t, err, "Failed to create in-memory test database")

	// Return cleanup function
	cleanup := func() {
		err := db.Close()
		require.NoError(t, err, "Failed to close in-memory test database")
	}

	return db, cleanup
}
