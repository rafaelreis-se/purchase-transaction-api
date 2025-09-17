package database_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/database"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExchangeRateRepository_Save(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())
	exchangeRate := fixtures.ValidExchangeRate()

	// Act
	err := repo.Save(&exchangeRate)

	// Assert
	assert.NoError(t, err)

	// Verify exchange rate was saved by fetching it back
	saved, err := repo.GetByID(exchangeRate.ID)
	require.NoError(t, err)
	require.NotNil(t, saved)

	assert.Equal(t, exchangeRate.ID, saved.ID)
	assert.Equal(t, exchangeRate.FromCurrency, saved.FromCurrency)
	assert.Equal(t, exchangeRate.ToCurrency, saved.ToCurrency)
	assert.Equal(t, exchangeRate.Rate, saved.Rate)
	assert.True(t, exchangeRate.EffectiveDate.Equal(saved.EffectiveDate))
}

func TestExchangeRateRepository_Save_Validation(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())

	t.Run("Nil exchange rate", func(t *testing.T) {
		err := repo.Save(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("Invalid exchange rate", func(t *testing.T) {
		// Create exchange rate with invalid data
		invalidRate := fixtures.ExchangeRateWithRate(-1.0) // Invalid: negative rate

		err := repo.Save(&invalidRate)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be positive")
	})
}

func TestExchangeRateRepository_GetByID(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())
	exchangeRate := fixtures.ValidExchangeRate()

	t.Run("Existing exchange rate", func(t *testing.T) {
		// Save exchange rate first
		err := repo.Save(&exchangeRate)
		require.NoError(t, err)

		// Act
		found, err := repo.GetByID(exchangeRate.ID)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, exchangeRate.ID, found.ID)
		assert.Equal(t, exchangeRate.FromCurrency, found.FromCurrency)
		assert.Equal(t, exchangeRate.ToCurrency, found.ToCurrency)
		assert.Equal(t, exchangeRate.Rate, found.Rate)
	})

	t.Run("Non-existing exchange rate", func(t *testing.T) {
		// Act
		randomID := uuid.New()
		found, err := repo.GetByID(randomID)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, found) // Should return nil when not found
	})
}

func TestExchangeRateRepository_FindRateForConversion(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())
	transactionDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	t.Run("Find valid rate within 6 months", func(t *testing.T) {
		// Create exchange rate 2 months before transaction date (valid)
		validDate := transactionDate.AddDate(0, -2, 0) // 2 months ago
		exchangeRate := fixtures.ExchangeRateWithDate(validDate)
		exchangeRate.FromCurrency = entities.USD
		exchangeRate.ToCurrency = entities.BRL

		// Save exchange rate
		require.NoError(t, repo.Save(&exchangeRate))

		// Act
		found, err := repo.FindRateForConversion(entities.USD, entities.BRL, transactionDate)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, exchangeRate.ID, found.ID)
		assert.Equal(t, entities.USD, found.FromCurrency)
		assert.Equal(t, entities.BRL, found.ToCurrency)
	})

	t.Run("Rate exactly 6 months old (boundary case)", func(t *testing.T) {
		// Create exchange rate exactly 6 months before transaction date
		boundaryDate := transactionDate.AddDate(0, -6, 0) // Exactly 6 months ago
		exchangeRate := fixtures.ExchangeRateWithDate(boundaryDate)
		exchangeRate.FromCurrency = entities.USD
		exchangeRate.ToCurrency = entities.EUR

		// Save exchange rate
		require.NoError(t, repo.Save(&exchangeRate))

		// Act
		found, err := repo.FindRateForConversion(entities.USD, entities.EUR, transactionDate)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, found) // Should find it (boundary included)
		assert.Equal(t, exchangeRate.ID, found.ID)
	})

	t.Run("Rate too old (more than 6 months)", func(t *testing.T) {
		// Create exchange rate 7 months before transaction date (too old)
		tooOldDate := transactionDate.AddDate(0, -7, 0) // 7 months ago
		exchangeRate := fixtures.ExchangeRateWithDate(tooOldDate)
		exchangeRate.FromCurrency = entities.USD
		exchangeRate.ToCurrency = entities.GBP

		// Save exchange rate
		require.NoError(t, repo.Save(&exchangeRate))

		// Act
		found, err := repo.FindRateForConversion(entities.USD, entities.GBP, transactionDate)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, found) // Should not find it (too old)
	})

	t.Run("Rate from future (after transaction date)", func(t *testing.T) {
		// Create exchange rate 1 day after transaction date (future)
		futureDate := transactionDate.AddDate(0, 0, 1) // 1 day later
		exchangeRate := fixtures.ExchangeRateWithDate(futureDate)
		exchangeRate.FromCurrency = entities.USD
		exchangeRate.ToCurrency = entities.JPY

		// Save exchange rate
		require.NoError(t, repo.Save(&exchangeRate))

		// Act
		found, err := repo.FindRateForConversion(entities.USD, entities.JPY, transactionDate)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, found) // Should not find it (from future)
	})

	t.Run("Multiple rates - returns most recent", func(t *testing.T) {
		// Create multiple valid exchange rates for same currency pair
		olderDate := transactionDate.AddDate(0, -3, 0) // 3 months ago
		newerDate := transactionDate.AddDate(0, -1, 0) // 1 month ago

		olderRate := fixtures.ExchangeRateWithDate(olderDate)
		olderRate.FromCurrency = entities.USD
		olderRate.ToCurrency = entities.CAD
		olderRate.Rate = 1.25

		newerRate := fixtures.ExchangeRateWithDate(newerDate)
		newerRate.FromCurrency = entities.USD
		newerRate.ToCurrency = entities.CAD
		newerRate.Rate = 1.30

		// Save both rates
		require.NoError(t, repo.Save(&olderRate))
		require.NoError(t, repo.Save(&newerRate))

		// Act
		found, err := repo.FindRateForConversion(entities.USD, entities.CAD, transactionDate)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, newerRate.ID, found.ID) // Should return the newer rate
		assert.Equal(t, 1.30, found.Rate)
	})

	t.Run("No matching currency pair", func(t *testing.T) {
		// Create exchange rate for different currency pair
		exchangeRate := fixtures.ValidExchangeRate()
		exchangeRate.FromCurrency = entities.EUR
		exchangeRate.ToCurrency = entities.GBP

		// Save exchange rate
		require.NoError(t, repo.Save(&exchangeRate))

		// Act - search for different currency pair
		found, err := repo.FindRateForConversion(entities.USD, entities.AUD, transactionDate)

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, found) // Should not find it (different currencies)
	})
}

func TestExchangeRateRepository_Update(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())
	exchangeRate := fixtures.ValidExchangeRate()

	t.Run("Update existing exchange rate", func(t *testing.T) {
		// Save exchange rate first
		require.NoError(t, repo.Save(&exchangeRate))

		// Modify exchange rate
		exchangeRate.Rate = 5.75
		exchangeRate.ToCurrency = entities.EUR

		// Act
		err := repo.Update(&exchangeRate)

		// Assert
		assert.NoError(t, err)

		// Verify changes were saved
		updated, err := repo.GetByID(exchangeRate.ID)
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, 5.75, updated.Rate)
		assert.Equal(t, entities.EUR, updated.ToCurrency)
	})

	t.Run("Update non-existing exchange rate", func(t *testing.T) {
		// Create exchange rate that doesn't exist in DB
		nonExistentRate := fixtures.ValidExchangeRate()

		// Act
		err := repo.Update(&nonExistentRate)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestExchangeRateRepository_Delete(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())
	exchangeRate := fixtures.ValidExchangeRate()

	t.Run("Delete existing exchange rate", func(t *testing.T) {
		// Save exchange rate first
		require.NoError(t, repo.Save(&exchangeRate))

		// Verify it exists
		exists, err := repo.Exists(exchangeRate.ID)
		require.NoError(t, err)
		require.True(t, exists)

		// Act
		err = repo.Delete(exchangeRate.ID)

		// Assert
		assert.NoError(t, err)

		// Verify it was deleted
		exists, err = repo.Exists(exchangeRate.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("Delete non-existing exchange rate", func(t *testing.T) {
		// Act
		randomID := uuid.New()
		err := repo.Delete(randomID)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestExchangeRateRepository_Exists(t *testing.T) {
	// Setup
	db, cleanup := setupInMemoryTestDB(t)
	defer cleanup()

	repo := database.NewExchangeRateRepository(db.GetDB())
	exchangeRate := fixtures.ValidExchangeRate()

	t.Run("Existing exchange rate", func(t *testing.T) {
		// Save exchange rate first
		require.NoError(t, repo.Save(&exchangeRate))

		// Act
		exists, err := repo.Exists(exchangeRate.ID)

		// Assert
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("Non-existing exchange rate", func(t *testing.T) {
		// Act
		randomID := uuid.New()
		exists, err := repo.Exists(randomID)

		// Assert
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}
