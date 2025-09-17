package repositories

import (
	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// TransactionRepository defines the contract for transaction persistence operations
type TransactionRepository interface {

	// Save persists a transaction to the database
	// Returns error if the operation fails
	Save(transaction *entities.Transaction) error

	// GetByID retrieves a transaction by its unique identifier
	// Returns nil and no error if transaction is not found
	GetByID(id uuid.UUID) (*entities.Transaction, error)

	// GetAll retrieves all transactions from the database
	// Returns empty slice if no transactions exist
	GetAll() ([]entities.Transaction, error)

	// Update modifies an existing transaction in the database
	// Returns error if transaction doesn't exist or operation fails
	Update(transaction *entities.Transaction) error

	// Delete removes a transaction from the database by ID
	// Returns error if transaction doesn't exist or operation fails
	Delete(id uuid.UUID) error

	// Exists checks if a transaction with the given ID exists
	// Returns true if exists, false otherwise
	Exists(id uuid.UUID) (bool, error)

	// Count returns the total number of transactions in the database
	Count() (int64, error)
}
