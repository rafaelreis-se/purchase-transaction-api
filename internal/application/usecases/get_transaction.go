package usecases

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
)

// GetTransactionUseCase handles the business logic for retrieving transactions
type GetTransactionUseCase struct {
	transactionRepo repositories.TransactionRepository
}

// NewGetTransactionUseCase creates a new instance of GetTransactionUseCase
func NewGetTransactionUseCase(transactionRepo repositories.TransactionRepository) *GetTransactionUseCase {
	return &GetTransactionUseCase{
		transactionRepo: transactionRepo,
	}
}

// Execute retrieves a transaction by its ID
func (uc *GetTransactionUseCase) Execute(id uuid.UUID) (*dto.GetTransactionResponse, error) {
	// Validate input
	if err := uc.validateInput(id); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get transaction from repository
	transaction, err := uc.transactionRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transaction: %w", err)
	}

	// Check if transaction was found
	if transaction == nil {
		return nil, fmt.Errorf("transaction not found with id: %s", id.String())
	}

	// Convert entity to response DTO
	response := dto.NewGetTransactionResponse(transaction)

	return response, nil
}

// validateInput validates the input parameters
func (uc *GetTransactionUseCase) validateInput(id uuid.UUID) error {
	// Check if UUID is valid (not nil/empty)
	if id == uuid.Nil {
		return fmt.Errorf("transaction ID cannot be empty")
	}

	return nil
}
