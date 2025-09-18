package usecases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransactionUseCase_Execute(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockTransactionRepository)
	validator := validator.New()
	usecase := usecases.NewCreateTransactionUseCase(mockRepo, validator)

	t.Run("Successful transaction creation", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "Test Purchase",
			Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Amount:      99.99,
		}

		// Mock the repository Save method to succeed
		mockRepo.On("Save", mock.AnythingOfType("*entities.Transaction")).Return(nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "Test Purchase", response.Description)
		assert.True(t, request.Date.Equal(response.Date))
		assert.Equal(t, 99.99, response.Amount)
		assert.False(t, response.CreatedAt.IsZero())

		// Verify mock was called
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

	t.Run("Invalid request - empty description", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "", // Invalid: empty
			Date:        time.Now(),
			Amount:      99.99,
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid request - description too long", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "This description is way too long and exceeds the fifty character limit that we have set", // > 50 chars
			Date:        time.Now(),
			Amount:      99.99,
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid request - zero amount", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "Valid description",
			Date:        time.Now(),
			Amount:      0, // Invalid: not greater than 0
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid request - negative amount", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "Valid description",
			Date:        time.Now(),
			Amount:      -10.50, // Invalid: negative
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid request - zero date", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "Valid description",
			Date:        time.Time{}, // Invalid: zero time
			Amount:      99.99,
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Repository save error", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "Test Purchase",
			Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Amount:      99.99,
		}

		// Mock the repository Save method to return an error
		repositoryError := errors.New("database connection failed")
		mockRepo.On("Save", mock.AnythingOfType("*entities.Transaction")).Return(repositoryError).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to save transaction")
		assert.Contains(t, err.Error(), "database connection failed")

		// Verify mock was called
		mockRepo.AssertExpectations(t)
	})

	t.Run("Business validation error", func(t *testing.T) {
		// Arrange - Create a request that passes struct validation but fails business validation
		// This test ensures our business validation layer works
		request := &dto.CreateTransactionRequest{
			Description: "This description is way too long for the business rules and should fail validation because it exceeds fifty characters",
			Date:        time.Now(),
			Amount:      99.99,
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestCreateTransactionUseCase_Constructor(t *testing.T) {
	t.Run("Valid constructor", func(t *testing.T) {
		// Arrange
		mockRepo := new(mocks.MockTransactionRepository)
		validator := validator.New()

		// Act
		usecase := usecases.NewCreateTransactionUseCase(mockRepo, validator)

		// Assert
		assert.NotNil(t, usecase)
	})
}

func TestCreateTransactionUseCase_DTOConversion(t *testing.T) {
	t.Run("Request to Entity conversion", func(t *testing.T) {
		// Arrange
		request := &dto.CreateTransactionRequest{
			Description: "Test Purchase",
			Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Amount:      99.99,
		}

		// Act
		entity := request.ToEntity()

		// Assert
		assert.NotNil(t, entity)
		assert.NotEmpty(t, entity.ID) // UUID should be generated
		assert.Equal(t, request.Description, entity.Description)
		assert.True(t, request.Date.Equal(entity.Date))
		assert.Equal(t, entities.NewMoney(request.Amount), entity.Amount)
	})

	t.Run("Entity to Response conversion", func(t *testing.T) {
		// Arrange
		entity := &entities.Transaction{
			ID:          uuid.New(),
			Description: "Test Purchase",
			Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			Amount:      entities.NewMoney(99.99),
			CreatedAt:   time.Now(),
		}

		// Act
		response := dto.NewCreateTransactionResponse(entity)

		// Assert
		assert.NotNil(t, response)
		assert.Equal(t, entity.ID, response.ID)
		assert.Equal(t, entity.Description, response.Description)
		assert.True(t, entity.Date.Equal(response.Date))
		assert.Equal(t, entity.Amount.Dollars(), response.Amount)
		assert.True(t, entity.CreatedAt.Equal(response.CreatedAt))
	})
}
