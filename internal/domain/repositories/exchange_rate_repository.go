package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// ExchangeRateRepository defines the contract for exchange rate persistence operations
type ExchangeRateRepository interface {
	// Save persists an exchange rate to the database
	// Returns error if the operation fails
	Save(exchangeRate *entities.ExchangeRate) error

	// GetByID retrieves an exchange rate by its unique identifier
	// Returns nil and no error if exchange rate is not found
	GetByID(id uuid.UUID) (*entities.ExchangeRate, error)

	// FindRateForConversion finds the most suitable exchange rate for currency conversion
	// Must comply with the 6-month rule: rate date <= transaction date and within 6 months
	// Returns the most recent valid rate, or nil if no valid rate exists
	FindRateForConversion(from, to entities.CurrencyCode, transactionDate time.Time) (*entities.ExchangeRate, error)

	// Update modifies an existing exchange rate in the database
	// Returns error if exchange rate doesn't exist or operation fails
	Update(exchangeRate *entities.ExchangeRate) error

	// Delete removes an exchange rate from the database by ID
	// Returns error if exchange rate doesn't exist or operation fails
	Delete(id uuid.UUID) error

	// Exists checks if an exchange rate with the given ID exists
	// Returns true if exists, false otherwise
	Exists(id uuid.UUID) (bool, error)
}
