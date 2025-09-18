package usecases

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
)

// CreateTransactionUseCase handles the business logic for creating transactions
type CreateTransactionUseCase struct {
	transactionRepo repositories.TransactionRepository
	validator       *validator.Validate
}

// NewCreateTransactionUseCase creates a new instance of CreateTransactionUseCase
func NewCreateTransactionUseCase(
	transactionRepo repositories.TransactionRepository,
	validator *validator.Validate,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		transactionRepo: transactionRepo,
		validator:       validator,
	}
}

// Execute creates a new transaction with the provided request data
func (uc *CreateTransactionUseCase) Execute(request *dto.CreateTransactionRequest) (*dto.CreateTransactionResponse, error) {
	// Validate input
	if err := uc.validateRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Convert DTO to entity
	transaction := request.ToEntity()

	// Additional business validation (beyond struct tags)
	if err := uc.validateBusinessRules(transaction); err != nil {
		return nil, fmt.Errorf("business validation failed: %w", err)
	}

	// Save transaction to repository
	if err := uc.transactionRepo.Save(transaction); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	// Convert entity back to response DTO
	response := dto.NewCreateTransactionResponse(transaction)

	return response, nil
}

// validateRequest validates the input request using struct tags
func (uc *CreateTransactionUseCase) validateRequest(request *dto.CreateTransactionRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if err := uc.validator.Struct(request); err != nil {
		return err
	}

	return nil
}

// validateBusinessRules performs additional business logic validation
func (uc *CreateTransactionUseCase) validateBusinessRules(transaction *entities.Transaction) error {
	// Use the entity's built-in validation
	if err := transaction.Validate(); err != nil {
		return err
	}

	return nil
}
