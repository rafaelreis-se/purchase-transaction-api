package usecases

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
)

// ConvertTransactionUseCase handles the business logic for currency conversion of transactions
type ConvertTransactionUseCase struct {
	transactionRepo  repositories.TransactionRepository
	exchangeRateRepo repositories.ExchangeRateRepository
	validator        *validator.Validate
}

// NewConvertTransactionUseCase creates a new instance of ConvertTransactionUseCase
func NewConvertTransactionUseCase(
	transactionRepo repositories.TransactionRepository,
	exchangeRateRepo repositories.ExchangeRateRepository,
	validator *validator.Validate,
) *ConvertTransactionUseCase {
	return &ConvertTransactionUseCase{
		transactionRepo:  transactionRepo,
		exchangeRateRepo: exchangeRateRepo,
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
func (uc *ConvertTransactionUseCase) validateConversionRules(_ *entities.Transaction, targetCurrency entities.CurrencyCode) error {
	// Validate target currency
	if !targetCurrency.IsValid() {
		return fmt.Errorf("invalid target currency: %s", targetCurrency)
	}

	// Check if conversion from USD to same currency (should be USD originally)
	if targetCurrency == entities.USD {
		return fmt.Errorf("cannot convert USD transaction to USD")
	}

	return nil
}

// findExchangeRate finds a suitable exchange rate implementing the 6-month rule
func (uc *ConvertTransactionUseCase) findExchangeRate(targetCurrency entities.CurrencyCode, transactionDate time.Time) (*entities.ExchangeRate, error) {
	// Find exchange rate from USD to target currency within 6 months of transaction date
	exchangeRate, err := uc.exchangeRateRepo.FindRateForConversion(entities.USD, targetCurrency, transactionDate)
	if err != nil {
		return nil, err
	}

	if exchangeRate == nil {
		return nil, fmt.Errorf("no suitable exchange rate found for %s within 6 months of transaction date %s",
			targetCurrency, transactionDate.Format("2006-01-02"))
	}

	return exchangeRate, nil
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
