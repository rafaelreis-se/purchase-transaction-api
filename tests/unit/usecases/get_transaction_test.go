package usecases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionUseCase_Execute(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockTransactionRepository)
	usecase := usecases.NewGetTransactionUseCase(mockRepo)

	t.Run("Successful transaction retrieval", func(t *testing.T) {
		// Arrange
		transactionID := uuid.New()
		expectedTransaction := fixtures.ValidTransaction()
		expectedTransaction.ID = transactionID

		// Mock the repository GetByID method to return the transaction
		mockRepo.On("GetByID", transactionID).Return(&expectedTransaction, nil).Once()

		// Act
		response, err := usecase.Execute(transactionID)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, expectedTransaction.ID, response.ID)
		assert.Equal(t, expectedTransaction.Description, response.Description)
		assert.True(t, expectedTransaction.Date.Equal(response.Date))
		assert.Equal(t, expectedTransaction.Amount.Dollars(), response.Amount)
		assert.True(t, expectedTransaction.CreatedAt.Equal(response.CreatedAt))
		assert.True(t, expectedTransaction.UpdatedAt.Equal(response.UpdatedAt))

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	t.Run("Transaction not found", func(t *testing.T) {
		// Arrange
		transactionID := uuid.New()

		// Mock the repository GetByID method to return nil (not found)
		mockRepo.On("GetByID", transactionID).Return(nil, nil).Once()

		// Act
		response, err := usecase.Execute(transactionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "transaction not found")
		assert.Contains(t, err.Error(), transactionID.String())

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	t.Run("Repository error", func(t *testing.T) {
		// Arrange
		transactionID := uuid.New()
		repositoryError := errors.New("database connection failed")

		// Mock the repository GetByID method to return an error
		mockRepo.On("GetByID", transactionID).Return(nil, repositoryError).Once()

		// Act
		response, err := usecase.Execute(transactionID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to retrieve transaction")
		assert.Contains(t, err.Error(), "database connection failed")

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid UUID - nil UUID", func(t *testing.T) {
		// Arrange
		nilUUID := uuid.Nil

		// Act
		response, err := usecase.Execute(nilUUID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "ID cannot be empty")

		// Repository should not be called for invalid input
		mockRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("Valid UUID format", func(t *testing.T) {
		// Arrange
		validUUID := uuid.New()
		transaction := fixtures.ValidTransaction()
		transaction.ID = validUUID

		// Mock successful repository call
		mockRepo.On("GetByID", validUUID).Return(&transaction, nil).Once()

		// Act
		response, err := usecase.Execute(validUUID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, validUUID, response.ID)

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})
}

func TestGetTransactionUseCase_Constructor(t *testing.T) {
	t.Run("Valid constructor", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockTransactionRepository)

		// Act
		usecase := usecases.NewGetTransactionUseCase(mockRepo)

		// Assert
		assert.NotNil(t, usecase)
	})

	t.Run("Constructor with nil repository", func(t *testing.T) {
		// Act
		usecase := usecases.NewGetTransactionUseCase(nil)

		// Assert
		assert.NotNil(t, usecase) // Constructor doesn't validate, but usecase will fail at runtime
	})
}

func TestGetTransactionUseCase_DTOConversion(t *testing.T) {
	t.Run("Entity to GetTransactionResponse conversion", func(t *testing.T) {
		// Arrange
		entity := &entities.Transaction{
			ID:          uuid.New(),
			Description: "Test Purchase",
			Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Amount:      entities.NewMoney(99.99),
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2024, 1, 15, 10, 35, 0, 0, time.UTC),
		}

		// Act
		response := dto.NewGetTransactionResponse(entity)

		// Assert
		assert.NotNil(t, response)
		assert.Equal(t, entity.ID, response.ID)
		assert.Equal(t, entity.Description, response.Description)
		assert.True(t, entity.Date.Equal(response.Date))
		assert.Equal(t, entity.Amount.Dollars(), response.Amount)
		assert.True(t, entity.CreatedAt.Equal(response.CreatedAt))
		assert.True(t, entity.UpdatedAt.Equal(response.UpdatedAt))
	})

	t.Run("Response contains all required fields", func(t *testing.T) {
		// Arrange
		entity := fixtures.ValidTransaction()

		// Act
		response := dto.NewGetTransactionResponse(&entity)

		// Assert
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.ID)
		assert.NotEmpty(t, response.Description)
		assert.False(t, response.Date.IsZero())
		assert.True(t, response.Amount > 0)
		assert.False(t, response.CreatedAt.IsZero())
		assert.False(t, response.UpdatedAt.IsZero())
	})
}
