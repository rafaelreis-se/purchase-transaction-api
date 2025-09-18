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
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertTransactionUseCase_Execute(t *testing.T) {
	// Setup
	mockTransactionRepo := new(mocks.MockTransactionRepository)
	mockExchangeRateRepo := new(mocks.MockExchangeRateRepository)
	validator := validator.New()
	usecase := usecases.NewConvertTransactionUseCase(mockTransactionRepo, mockExchangeRateRepo, validator)

	t.Run("Successful currency conversion", func(t *testing.T) {
		// Arrange
		transactionID := uuid.New()
		transaction := fixtures.ValidTransaction()
		transaction.ID = transactionID
		transaction.Date = time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

		exchangeRate := fixtures.ValidExchangeRate()
		exchangeRate.FromCurrency = entities.USD
		exchangeRate.ToCurrency = entities.BRL
		exchangeRate.Rate = 5.20
		exchangeRate.EffectiveDate = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // 5 days before transaction

		request := &dto.ConvertTransactionRequest{
			TransactionID:  transactionID,
			TargetCurrency: entities.BRL,
		}

		// Mock repository calls
		mockTransactionRepo.On("GetByID", transactionID).Return(&transaction, nil).Once()
		mockExchangeRateRepo.On("FindRateForConversion", entities.USD, entities.BRL, transaction.Date).Return(&exchangeRate, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, transaction.ID, response.Transaction.ID)
		assert.Equal(t, entities.BRL, response.TargetCurrency)
		assert.Equal(t, 5.20, response.ExchangeRate)
		assert.True(t, response.ConvertedAmount > 0)
		assert.Equal(t, exchangeRate.EffectiveDate, response.EffectiveDate)

		// Verify converted amount calculation
		expectedAmount := transaction.Amount.Dollars() * exchangeRate.Rate
		assert.InDelta(t, expectedAmount, response.ConvertedAmount, 0.01)

		// Verify mocks were called
		mockTransactionRepo.AssertExpectations(t)
		mockExchangeRateRepo.AssertExpectations(t)
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

	t.Run("Invalid request - empty transaction ID", func(t *testing.T) {
		// Arrange
		request := &dto.ConvertTransactionRequest{
			TransactionID:  uuid.Nil, // Invalid: empty UUID
			TargetCurrency: entities.BRL,
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Invalid request - empty target currency", func(t *testing.T) {
		// Arrange
		request := &dto.ConvertTransactionRequest{
			TransactionID:  uuid.New(),
			TargetCurrency: "", // Invalid: empty currency
		}

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Transaction not found", func(t *testing.T) {
		// Arrange
		request := &dto.ConvertTransactionRequest{
			TransactionID:  uuid.New(),
			TargetCurrency: entities.BRL,
		}

		// Mock transaction repository to return nil (not found)
		mockTransactionRepo.On("GetByID", request.TransactionID).Return(nil, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get transaction")
		assert.Contains(t, err.Error(), "transaction not found")

		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("Transaction repository error", func(t *testing.T) {
		// Arrange
		request := &dto.ConvertTransactionRequest{
			TransactionID:  uuid.New(),
			TargetCurrency: entities.BRL,
		}

		repositoryError := errors.New("database connection failed")
		mockTransactionRepo.On("GetByID", request.TransactionID).Return(nil, repositoryError).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get transaction")
		assert.Contains(t, err.Error(), "database connection failed")

		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("Invalid target currency - USD to USD conversion", func(t *testing.T) {
		// Arrange
		transaction := fixtures.ValidTransaction()
		request := &dto.ConvertTransactionRequest{
			TransactionID:  transaction.ID,
			TargetCurrency: entities.USD, // Invalid: cannot convert USD to USD
		}

		mockTransactionRepo.On("GetByID", request.TransactionID).Return(&transaction, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "conversion validation failed")
		assert.Contains(t, err.Error(), "cannot convert USD transaction to USD")

		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("No suitable exchange rate found", func(t *testing.T) {
		// Arrange
		transaction := fixtures.ValidTransaction()
		request := &dto.ConvertTransactionRequest{
			TransactionID:  transaction.ID,
			TargetCurrency: entities.BRL,
		}

		mockTransactionRepo.On("GetByID", request.TransactionID).Return(&transaction, nil).Once()
		// Mock exchange rate repository to return nil (no rate found)
		mockExchangeRateRepo.On("FindRateForConversion", entities.USD, entities.BRL, transaction.Date).Return(nil, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to find exchange rate")
		assert.Contains(t, err.Error(), "no suitable exchange rate found")
		assert.Contains(t, err.Error(), "within 6 months")

		mockTransactionRepo.AssertExpectations(t)
		mockExchangeRateRepo.AssertExpectations(t)
	})

	t.Run("Exchange rate repository error", func(t *testing.T) {
		// Arrange
		transaction := fixtures.ValidTransaction()
		request := &dto.ConvertTransactionRequest{
			TransactionID:  transaction.ID,
			TargetCurrency: entities.BRL,
		}

		repositoryError := errors.New("exchange rate service unavailable")
		mockTransactionRepo.On("GetByID", request.TransactionID).Return(&transaction, nil).Once()
		mockExchangeRateRepo.On("FindRateForConversion", entities.USD, entities.BRL, transaction.Date).Return(nil, repositoryError).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to find exchange rate")
		assert.Contains(t, err.Error(), "exchange rate service unavailable")

		mockTransactionRepo.AssertExpectations(t)
		mockExchangeRateRepo.AssertExpectations(t)
	})

	t.Run("Invalid exchange rate - date validation fails", func(t *testing.T) {
		// Arrange
		transaction := fixtures.ValidTransaction()
		transaction.Date = time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC)

		// Create exchange rate with invalid date (too old for the 6-month rule)
		invalidExchangeRate := fixtures.ValidExchangeRate()
		invalidExchangeRate.FromCurrency = entities.USD
		invalidExchangeRate.ToCurrency = entities.BRL
		invalidExchangeRate.EffectiveDate = time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC) // More than 6 months ago

		request := &dto.ConvertTransactionRequest{
			TransactionID:  transaction.ID,
			TargetCurrency: entities.BRL,
		}

		mockTransactionRepo.On("GetByID", request.TransactionID).Return(&transaction, nil).Once()
		mockExchangeRateRepo.On("FindRateForConversion", entities.USD, entities.BRL, transaction.Date).Return(&invalidExchangeRate, nil).Once()

		// Act
		response, err := usecase.Execute(request)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to create converted transaction")
		assert.Contains(t, err.Error(), "not within 6 months")

		mockTransactionRepo.AssertExpectations(t)
		mockExchangeRateRepo.AssertExpectations(t)
	})

	t.Run("Different target currencies", func(t *testing.T) {
		// Test multiple currencies to ensure the use case works for different targets
		testCases := []struct {
			name           string
			targetCurrency entities.CurrencyCode
			exchangeRate   float64
		}{
			{"USD to EUR", entities.EUR, 0.85},
			{"USD to GBP", entities.GBP, 0.75},
			{"USD to JPY", entities.JPY, 110.0},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Arrange
				transaction := fixtures.ValidTransaction()
				transaction.Amount = entities.NewMoney(100.00) // $100 for easy calculation

				exchangeRate := fixtures.ValidExchangeRate()
				exchangeRate.FromCurrency = entities.USD
				exchangeRate.ToCurrency = tc.targetCurrency
				exchangeRate.Rate = tc.exchangeRate

				request := &dto.ConvertTransactionRequest{
					TransactionID:  transaction.ID,
					TargetCurrency: tc.targetCurrency,
				}

				mockTransactionRepo.On("GetByID", request.TransactionID).Return(&transaction, nil).Once()
				mockExchangeRateRepo.On("FindRateForConversion", entities.USD, tc.targetCurrency, transaction.Date).Return(&exchangeRate, nil).Once()

				// Act
				response, err := usecase.Execute(request)

				// Assert
				assert.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tc.targetCurrency, response.TargetCurrency)
				assert.Equal(t, tc.exchangeRate, response.ExchangeRate)

				// Verify amount calculation
				expectedAmount := 100.00 * tc.exchangeRate
				assert.InDelta(t, expectedAmount, response.ConvertedAmount, 0.01)

				mockTransactionRepo.AssertExpectations(t)
				mockExchangeRateRepo.AssertExpectations(t)
			})
		}
	})
}

func TestConvertTransactionUseCase_Constructor(t *testing.T) {
	t.Run("Valid constructor", func(t *testing.T) {
		// Arrange
		mockTransactionRepo := new(mocks.MockTransactionRepository)
		mockExchangeRateRepo := new(mocks.MockExchangeRateRepository)
		validator := validator.New()

		// Act
		usecase := usecases.NewConvertTransactionUseCase(mockTransactionRepo, mockExchangeRateRepo, validator)

		// Assert
		assert.NotNil(t, usecase)
	})
}
