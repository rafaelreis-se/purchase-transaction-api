package usecases

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/repositories"
)

// ListTransactionsUseCase handles the business logic for listing transactions with pagination
type ListTransactionsUseCase struct {
	transactionRepo repositories.TransactionRepository
	validator       *validator.Validate
}

// NewListTransactionsUseCase creates a new instance of ListTransactionsUseCase
func NewListTransactionsUseCase(
	transactionRepo repositories.TransactionRepository,
	validator *validator.Validate,
) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{
		transactionRepo: transactionRepo,
		validator:       validator,
	}
}

// Execute retrieves a paginated list of transactions
func (uc *ListTransactionsUseCase) Execute(request *dto.ListTransactionsRequest) (*dto.ListTransactionsResponse, error) {
	// Validate and set defaults for request
	if err := uc.validateAndSetDefaults(request); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get paginated transactions from repository
	transactions, total, err := uc.transactionRepo.GetAllPaginated(request.Page, request.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}

	// Convert to response DTO with pagination metadata
	response := dto.NewListTransactionsResponse(transactions, request.Page, request.Size, total)

	return response, nil
}

// validateAndSetDefaults validates the request and sets default values
func (uc *ListTransactionsUseCase) validateAndSetDefaults(request *dto.ListTransactionsRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Set defaults if not provided
	if request.Page == 0 {
		request.Page = 1
	}
	if request.Size == 0 {
		request.Size = 20
	}

	// Validate constraints
	if request.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}
	if request.Size < 1 {
		return fmt.Errorf("size must be at least 1")
	}
	if request.Size > 100 {
		return fmt.Errorf("size cannot exceed 100")
	}

	// Use validator for struct validation
	if err := uc.validator.Struct(request); err != nil {
		return err
	}

	return nil
}
