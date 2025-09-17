package database_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/database"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRepository_Save(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())
	transaction := fixtures.ValidTransaction()

	// Act
	err := repo.Save(&transaction)

	// Assert
	assert.NoError(t, err)

	// Verify transaction was saved by fetching it back
	saved, err := repo.GetByID(transaction.ID)
	require.NoError(t, err)
	require.NotNil(t, saved)

	assert.Equal(t, transaction.ID, saved.ID)
	assert.Equal(t, transaction.Description, saved.Description)
	assert.Equal(t, transaction.Amount, saved.Amount)
	assert.True(t, transaction.Date.Equal(saved.Date))
}

func TestTransactionRepository_Save_Validation(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())

	t.Run("Nil transaction", func(t *testing.T) {
		err := repo.Save(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("Invalid transaction", func(t *testing.T) {
		// Create transaction with invalid data
		invalidTx := fixtures.ValidTransaction()
		invalidTx.Description = "" // Invalid: empty description

		err := repo.Save(&invalidTx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "description is required")
	})
}

func TestTransactionRepository_GetByID(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())
	transaction := fixtures.ValidTransaction()

	t.Run("Existing transaction", func(t *testing.T) {
		// Save transaction first
		err := repo.Save(&transaction)
		require.NoError(t, err)

		// Act
		found, err := repo.GetByID(transaction.ID)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, transaction.ID, found.ID)
		assert.Equal(t, transaction.Description, found.Description)
	})

	t.Run("Non-existing transaction", func(t *testing.T) {
		// Act
		randomID := uuid.New()
		found, err := repo.GetByID(randomID)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, found) // Should return nil when not found
	})
}

func TestTransactionRepository_GetAll(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())

	t.Run("Empty database", func(t *testing.T) {
		// Act
		transactions, err := repo.GetAll()

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, transactions)
	})

	t.Run("Multiple transactions", func(t *testing.T) {
		// Save multiple transactions
		tx1 := fixtures.ValidTransaction()
		tx2 := fixtures.TransactionWithDescription("Another transaction")
		tx3 := fixtures.TransactionWithAmount(25.50)

		require.NoError(t, repo.Save(&tx1))
		require.NoError(t, repo.Save(&tx2))
		require.NoError(t, repo.Save(&tx3))

		// Act
		transactions, err := repo.GetAll()

		// Assert
		assert.NoError(t, err)
		assert.Len(t, transactions, 3)

		// Check that all transactions are present
		ids := make(map[uuid.UUID]bool)
		for _, tx := range transactions {
			ids[tx.ID] = true
		}
		assert.True(t, ids[tx1.ID])
		assert.True(t, ids[tx2.ID])
		assert.True(t, ids[tx3.ID])
	})
}

func TestTransactionRepository_Update(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())
	transaction := fixtures.ValidTransaction()

	t.Run("Update existing transaction", func(t *testing.T) {
		// Save transaction first
		require.NoError(t, repo.Save(&transaction))

		// Modify transaction
		transaction.Description = "Updated description"
		transaction.Amount = entities.NewMoney(150.75)

		// Act
		err := repo.Update(&transaction)

		// Assert
		assert.NoError(t, err)

		// Verify changes were saved
		updated, err := repo.GetByID(transaction.ID)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, entities.NewMoney(150.75), updated.Amount)
	})

	t.Run("Update non-existing transaction", func(t *testing.T) {
		// Create transaction that doesn't exist in DB
		nonExistentTx := fixtures.ValidTransaction()

		// Act
		err := repo.Update(&nonExistentTx)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTransactionRepository_Delete(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())
	transaction := fixtures.ValidTransaction()

	t.Run("Delete existing transaction", func(t *testing.T) {
		// Save transaction first
		require.NoError(t, repo.Save(&transaction))

		// Verify it exists
		exists, err := repo.Exists(transaction.ID)
		require.NoError(t, err)
		require.True(t, exists)

		// Act
		err = repo.Delete(transaction.ID)

		// Assert
		assert.NoError(t, err)

		// Verify it was deleted
		exists, err = repo.Exists(transaction.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Delete non-existing transaction", func(t *testing.T) {
		// Act
		randomID := uuid.New()
		err := repo.Delete(randomID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestTransactionRepository_Exists(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())
	transaction := fixtures.ValidTransaction()

	t.Run("Existing transaction", func(t *testing.T) {
		// Save transaction first
		require.NoError(t, repo.Save(&transaction))

		// Act
		exists, err := repo.Exists(transaction.ID)

		// Assert
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Non-existing transaction", func(t *testing.T) {
		// Act
		randomID := uuid.New()
		exists, err := repo.Exists(randomID)

		// Assert
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestTransactionRepository_Count(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewTransactionRepository(db.GetDB())

	t.Run("Empty database", func(t *testing.T) {
		// Act
		count, err := repo.Count()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("Multiple transactions", func(t *testing.T) {
		// Save multiple transactions
		tx1 := fixtures.ValidTransaction()
		tx2 := fixtures.ValidTransaction()
		tx3 := fixtures.ValidTransaction()

		require.NoError(t, repo.Save(&tx1))
		require.NoError(t, repo.Save(&tx2))
		require.NoError(t, repo.Save(&tx3))

		// Act
		count, err := repo.Count()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})
}
