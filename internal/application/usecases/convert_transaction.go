package usecases

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/services"
)

// ConvertTransactionUseCase handles the business logic for currency conversion of transactions
type ConvertTransactionUseCase struct {
	transactionRepo  repositories.TransactionRepository
	exchangeRateRepo repositories.ExchangeRateRepository
	treasuryService  services.TreasuryService
	validator        *validator.Validate
}

// NewConvertTransactionUseCase creates a new instance of ConvertTransactionUseCase
func NewConvertTransactionUseCase(
	transactionRepo repositories.TransactionRepository,
	exchangeRateRepo repositories.ExchangeRateRepository,
	treasuryService services.TreasuryService,
	validator *validator.Validate,
) *ConvertTransactionUseCase {
	return &ConvertTransactionUseCase{
		transactionRepo:  transactionRepo,
		exchangeRateRepo: exchangeRateRepo,
		treasuryService:  treasuryService,
		validator:        validator,
	}
}

// Execute converts a transaction to the specified target currency
func (uc *ConvertTransactionUseCase) Execute(request *dto.ConvertTransactionRequest) (*dto.ConvertTransactionResponse, error) {
	// Validate input request
	if err := uc.validateRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get the original transaction
	transaction, err := uc.getTransaction(request.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Validate business rules for conversion
	if err := uc.validateConversionRules(transaction, request.TargetCurrency); err != nil {
		return nil, fmt.Errorf("conversion validation failed: %w", err)
	}

	// Find suitable exchange rate (implements 6-month rule)
	exchangeRate, err := uc.findExchangeRate(request.TargetCurrency, transaction.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to find exchange rate: %w", err)
	}

	// Create converted transaction with the found exchange rate
	convertedTransaction, err := uc.createConvertedTransaction(transaction, request.TargetCurrency, exchangeRate)
	if err != nil {
		return nil, fmt.Errorf("failed to create converted transaction: %w", err)
	}

	// Convert to response DTO
	response := dto.NewConvertTransactionResponse(convertedTransaction)

	return response, nil
}

// validateRequest validates the input request using struct tags
func (uc *ConvertTransactionUseCase) validateRequest(request *dto.ConvertTransactionRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if err := uc.validator.Struct(request); err != nil {
		return err
	}

	return nil
}

// getTransaction retrieves the transaction by ID
func (uc *ConvertTransactionUseCase) getTransaction(transactionID uuid.UUID) (*entities.Transaction, error) {
	transaction, err := uc.transactionRepo.GetByID(transactionID)
	if err != nil {
		return nil, err
	}

	if transaction == nil {
		return nil, fmt.Errorf("transaction not found with id: %s", transactionID.String())
	}

	return transaction, nil
}

// validateConversionRules validates business rules for currency conversion
func (uc *ConvertTransactionUseCase) validateConversionRules(transaction *entities.Transaction, targetCurrency entities.CurrencyCode) error {
	// Validate target currency
	if !targetCurrency.IsValid() {
		return fmt.Errorf("invalid target currency: %s", targetCurrency)
	}

	// Check if conversion from USD to same currency (should be USD originally)
	if targetCurrency == entities.USD {
		return fmt.Errorf("cannot convert USD transaction to USD")
	}

	// Additional business rules can be added here
	// For example: check if transaction is not too old, business hours, etc.

	return nil
}

// findExchangeRate finds a suitable exchange rate implementing the 6-month rule
// First tries local repository, then falls back to Treasury API
func (uc *ConvertTransactionUseCase) findExchangeRate(targetCurrency entities.CurrencyCode, transactionDate time.Time) (*entities.ExchangeRate, error) {
	// 1. First, try to find exchange rate in local repository
	exchangeRate, err := uc.exchangeRateRepo.FindRateForConversion(entities.USD, targetCurrency, transactionDate)
	if err != nil {
		return nil, fmt.Errorf("error searching local exchange rates: %w", err)
	}

	// 2. If found in local repository, return it
	if exchangeRate != nil {
		return exchangeRate, nil
	}

	// 3. If not found locally, fetch from Treasury API
	treasuryRate, err := uc.treasuryService.FetchExchangeRate(entities.USD, targetCurrency, transactionDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate from Treasury API: %w", err)
	}

	// 4. Save the fetched rate to local repository for future use (caching)
	if err := uc.exchangeRateRepo.Save(treasuryRate); err != nil {
		// Log error but don't fail the conversion - we still have the rate
		slog.Warn("Failed to cache exchange rate from Treasury API",
			"error", err.Error(),
			"from_currency", string(entities.USD),
			"to_currency", string(targetCurrency),
			"rate", treasuryRate.Rate,
		)
	}

	return treasuryRate, nil
}

// createConvertedTransaction creates a ConvertedTransaction entity with validation
func (uc *ConvertTransactionUseCase) createConvertedTransaction(
	transaction *entities.Transaction,
	targetCurrency entities.CurrencyCode,
	exchangeRate *entities.ExchangeRate,
) (*entities.ConvertedTransaction, error) {
	// Use the entity's factory method which includes validation
	convertedTransaction, err := entities.NewConvertedTransaction(*transaction, targetCurrency, exchangeRate)
	if err != nil {
		return nil, err
	}

	return convertedTransaction, nil
}
