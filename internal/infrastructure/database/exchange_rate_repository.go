package database

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// sqliteExchangeRateRepository implements ExchangeRateRepository interface using SQLite
type sqliteExchangeRateRepository struct {
	db *gorm.DB
}

// NewExchangeRateRepository creates a new SQLite implementation of ExchangeRateRepository
func NewExchangeRateRepository(db *gorm.DB) repositories.ExchangeRateRepository {
	return &sqliteExchangeRateRepository{
		db: db,
	}
}

// Save persists an exchange rate to the database
func (r *sqliteExchangeRateRepository) Save(exchangeRate *entities.ExchangeRate) error {
	if exchangeRate == nil {
		return errors.New("exchange rate cannot be nil")
	}

	// Validate exchange rate before saving
	if err := exchangeRate.Validate(); err != nil {
		return err
	}

	// Create exchange rate in database
	result := r.db.Create(exchangeRate)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetByID retrieves an exchange rate by its unique identifier
func (r *sqliteExchangeRateRepository) GetByID(id uuid.UUID) (*entities.ExchangeRate, error) {
	var exchangeRate entities.ExchangeRate

	result := r.db.First(&exchangeRate, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil when not found (as per interface contract)
		}
		return nil, result.Error
	}

	return &exchangeRate, nil
}

// FindRateForConversion finds the most suitable exchange rate for currency conversion
// Must comply with the 6-month rule: rate date <= transaction date and within 6 months
func (r *sqliteExchangeRateRepository) FindRateForConversion(from, to entities.CurrencyCode, transactionDate time.Time) (*entities.ExchangeRate, error) {
	// Calculate 6 months ago from transaction date
	sixMonthsAgo := transactionDate.AddDate(0, -6, 0)

	var exchangeRate entities.ExchangeRate

	// Find the most recent exchange rate that satisfies the 6-month rule
	result := r.db.Where("from_currency = ? AND to_currency = ?", from, to).
		Where("effective_date <= ?", transactionDate). // Rate date <= transaction date
		Where("effective_date >= ?", sixMonthsAgo).    // Within 6 months
		Order("effective_date DESC").                  // Most recent first
		First(&exchangeRate)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No suitable rate found
		}
		return nil, result.Error
	}

	return &exchangeRate, nil
}

// Update modifies an existing exchange rate in the database
func (r *sqliteExchangeRateRepository) Update(exchangeRate *entities.ExchangeRate) error {
	if exchangeRate == nil {
		return errors.New("exchange rate cannot be nil")
	}

	// Validate exchange rate before updating
	if err := exchangeRate.Validate(); err != nil {
		return err
	}

	// Check if exchange rate exists
	exists, err := r.Exists(exchangeRate.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("exchange rate not found")
	}

	// Update exchange rate in database
	result := r.db.Save(exchangeRate)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Delete removes an exchange rate from the database by ID
func (r *sqliteExchangeRateRepository) Delete(id uuid.UUID) error {
	// Check if exchange rate exists
	exists, err := r.Exists(id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("exchange rate not found")
	}

	// Delete exchange rate from database
	result := r.db.Delete(&entities.ExchangeRate{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Exists checks if an exchange rate with the given ID exists
func (r *sqliteExchangeRateRepository) Exists(id uuid.UUID) (bool, error) {
	var count int64

	result := r.db.Model(&entities.ExchangeRate{}).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}
