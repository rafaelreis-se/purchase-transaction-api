package database

import (
	"errors"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// sqliteTransactionRepository implements TransactionRepository interface using SQLite
type sqliteTransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new SQLite implementation of TransactionRepository
func NewTransactionRepository(db *gorm.DB) repositories.TransactionRepository {
	return &sqliteTransactionRepository{
		db: db,
	}
}

// Save persists a transaction to the database
func (r *sqliteTransactionRepository) Save(transaction *entities.Transaction) error {
	if transaction == nil {
		return errors.New("transaction cannot be nil")
	}

	// Validate transaction before saving
	if err := transaction.Validate(); err != nil {
		return err
	}

	// Create transaction in database
	result := r.db.Create(transaction)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetByID retrieves a transaction by its unique identifier
func (r *sqliteTransactionRepository) GetByID(id uuid.UUID) (*entities.Transaction, error) {
	var transaction entities.Transaction

	result := r.db.First(&transaction, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil when not found (as per interface contract)
		}
		return nil, result.Error
	}

	return &transaction, nil
}

// GetAll retrieves all transactions from the database
func (r *sqliteTransactionRepository) GetAll() ([]entities.Transaction, error) {
	var transactions []entities.Transaction

	result := r.db.Find(&transactions)
	if result.Error != nil {
		return nil, result.Error
	}

	return transactions, nil
}

// GetAllPaginated retrieves transactions with pagination support
func (r *sqliteTransactionRepository) GetAllPaginated(page, size int) ([]entities.Transaction, int64, error) {
	var transactions []entities.Transaction
	var total int64

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20 // Default size
	}

	// Calculate offset
	offset := (page - 1) * size

	// Get total count
	result := r.db.Model(&entities.Transaction{}).Count(&total)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Get paginated transactions ordered by created_at DESC (most recent first)
	result = r.db.Order("created_at DESC").Limit(size).Offset(offset).Find(&transactions)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return transactions, total, nil
}

// Update modifies an existing transaction in the database
func (r *sqliteTransactionRepository) Update(transaction *entities.Transaction) error {
	if transaction == nil {
		return errors.New("transaction cannot be nil")
	}

	// Validate transaction before updating
	if err := transaction.Validate(); err != nil {
		return err
	}

	// Check if transaction exists
	exists, err := r.Exists(transaction.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("transaction not found")
	}

	// Update transaction in database
	result := r.db.Save(transaction)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Delete removes a transaction from the database by ID
func (r *sqliteTransactionRepository) Delete(id uuid.UUID) error {
	// Check if transaction exists
	exists, err := r.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("transaction not found")
	}

	// Delete transaction from database
	result := r.db.Delete(&entities.Transaction{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Exists checks if a transaction with the given ID exists
func (r *sqliteTransactionRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64

	result := r.db.Model(&entities.Transaction{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// Count returns the total number of transactions in the database
func (r *sqliteTransactionRepository) Count() (int64, error) {
	var count int64

	result := r.db.Model(&entities.Transaction{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}

	return count, nil
}
