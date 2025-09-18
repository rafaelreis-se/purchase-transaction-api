package usecases_test

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListTransactionsUseCase_Execute(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockTransactionRepository)
	validator := validator.New()
	usecase := usecases.NewListTransactionsUseCase(mockRepo, validator)

	t.Run("Successful pagination - first page", func(t *testing.T) {
		// Arrange
		transactions := []entities.Transaction{
			fixtures.ValidTransaction(),
			fixtures.TransactionWithDescription("Second transaction"),
			fixtures.TransactionWithAmount(25.50),
		}
		total := int64(10) // Total of 10 transactions in the system

		request := &dto.ListTransactionsRequest{
			Page: 1,
			Size: 20,
		}

		// Mock the repository GetAllPaginated method
		mockRepo.On("GetAllPaginated", 1, 20).Return(transactions, total, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.Len(t, response.Data, 3)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.Size)
		assert.Equal(t, int64(10), response.Total)
		assert.Equal(t, 1, response.TotalPages) // 10 items / 20 per page = 1 page

		// Verify transaction data
		assert.Equal(t, transactions[0].ID, response.Data[0].ID)
		assert.Equal(t, transactions[1].Description, response.Data[1].Description)
		assert.Equal(t, transactions[2].Amount.Dollars(), response.Data[2].Amount)

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	t.Run("Successful pagination - second page", func(t *testing.T) {
		// Arrange
		transactions := []entities.Transaction{
			fixtures.ValidTransaction(),
			fixtures.TransactionWithDescription("Another transaction"),
		}
		total := int64(22) // Total of 22 transactions

		request := &dto.ListTransactionsRequest{
			Page: 2,
			Size: 10,
		}

		mockRepo.On("GetAllPaginated", 2, 10).Return(transactions, total, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.Len(t, response.Data, 2)
		assert.Equal(t, 2, response.Page)
		assert.Equal(t, 10, response.Size)
		assert.Equal(t, int64(22), response.Total)
		assert.Equal(t, 3, response.TotalPages) // 22 items / 10 per page = 3 pages (ceiling)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty result set", func(t *testing.T) {
		// Arrange
		emptyTransactions := []entities.Transaction{}
		total := int64(0)

		request := &dto.ListTransactionsRequest{
			Page: 1,
			Size: 20,
		}

		mockRepo.On("GetAllPaginated", 1, 20).Return(emptyTransactions, total, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.Empty(t, response.Data)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.Size)
		assert.Equal(t, int64(0), response.Total)
		assert.Equal(t, 0, response.TotalPages)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Default values when request has zeros", func(t *testing.T) {
		// Arrange
		transactions := []entities.Transaction{fixtures.ValidTransaction()}
		total := int64(1)

		request := &dto.ListTransactionsRequest{
			Page: 0, // Should default to 1
			Size: 0, // Should default to 20
		}

		// Mock should be called with default values
		mockRepo.On("GetAllPaginated", 1, 20).Return(transactions, total, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 20, response.Size)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Nil request", func(t *testing.T) {
		// Act
		response, err := usecase.Execute(nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("Invalid page - negative", func(t *testing.T) {
		// Arrange
		request := &dto.ListTransactionsRequest{
			Page: -1, // Invalid
			Size: 20,
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "page must be at least 1")
	})

	t.Run("Invalid size - negative", func(t *testing.T) {
		// Arrange
		request := &dto.ListTransactionsRequest{
			Page: 1,
			Size: -5, // Invalid
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "size must be at least 1")
	})

	t.Run("Invalid size - too large", func(t *testing.T) {
		// Arrange
		request := &dto.ListTransactionsRequest{
			Page: 1,
			Size: 150, // Invalid - exceeds max 100
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "cannot exceed 100")
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		request := &dto.ListTransactionsRequest{
			Page: 1,
			Size: 20,
		}

		repositoryError := errors.New("database connection failed")
		mockRepo.On("GetAllPaginated", 1, 20).Return(nil, int64(0), repositoryError).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to retrieve transactions")
		assert.Contains(t, err.Error(), "database connection failed")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Different page sizes", func(t *testing.T) {
		// Test different valid page sizes
		testCases := []struct {
			name     string
			page     int
			size     int
			expected struct {
				page int
				size int
			}
		}{
			{"Small page size", 1, 5, struct{ page, size int }{1, 5}},
			{"Medium page size", 2, 50, struct{ page, size int }{2, 50}},
			{"Max page size", 1, 100, struct{ page, size int }{1, 100}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				transactions := []entities.Transaction{fixtures.ValidTransaction()}
				total := int64(1)

				request := &dto.ListTransactionsRequest{
					Page: tc.page,
					Size: tc.size,
				}

				mockRepo.On("GetAllPaginated", tc.expected.page, tc.expected.size).Return(transactions, total, nil).Once()

				// Act
				response, err := usecase.Execute(request)

				// Assert
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tc.expected.page, response.Page)
				assert.Equal(t, tc.expected.size, response.Size)

				mockRepo.AssertExpectations(t)
			})
		}
	})

	t.Run("Total pages calculation", func(t *testing.T) {
		// Test ceiling division for total pages calculation
		testCases := []struct {
			name          string
			total         int64
			size          int
			expectedPages int
		}{
			{"Exact division", 20, 10, 2},
			{"Ceiling needed", 21, 10, 3},
			{"Single page", 5, 10, 1},
			{"Empty", 0, 10, 0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				transactions := []entities.Transaction{} // Don't care about content
				request := &dto.ListTransactionsRequest{Page: 1, Size: tc.size}

				mockRepo.On("GetAllPaginated", 1, tc.size).Return(transactions, tc.total, nil).Once()

				// Act
				response, err := usecase.Execute(request)

				// Assert
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tc.expectedPages, response.TotalPages)
				assert.Equal(t, tc.total, response.Total)

				mockRepo.AssertExpectations(t)
			})
		}
	})
}

func TestListTransactionsUseCase_Constructor(t *testing.T) {
	t.Run("Valid constructor", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockTransactionRepository)
		validator := validator.New()

		// Act
		usecase := usecases.NewListTransactionsUseCase(mockRepo, validator)

		// Assert
		assert.NotNil(t, usecase)
	})
}
