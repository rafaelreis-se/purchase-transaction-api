package entities_test

import (
	"testing"
	"time"

	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestCurrencyCodeValidation(t *testing.T) {
	testCases := fixtures.CurrencyCodeTestCases()

	// Test valid currency codes
	for _, tc := range testCases.ValidCodes {
		t.Run("Valid_"+tc.Input, func(t *testing.T) {
			result, err := entities.NewCurrencyCode(tc.Input)

			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, result)
			assert.True(t, result.IsValid())
		})
	}

	// Test invalid currency codes
	for _, tc := range testCases.InvalidCodes {
		t.Run("Invalid_"+tc.Input, func(t *testing.T) {
			result, err := entities.NewCurrencyCode(tc.Input)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.ExpectedErr)
			assert.Equal(t, entities.CurrencyCode(""), result)
		})
	}
}

func TestCurrencyCodeString(t *testing.T) {
	tests := []struct {
		name     string
		currency entities.CurrencyCode
		expected string
	}{
		{"USD", entities.USD, "USD"},
		{"EUR", entities.EUR, "EUR"},
		{"BRL", entities.BRL, "BRL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.currency.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExchangeRateValidation(t *testing.T) {
	validationCases := fixtures.ExchangeRateValidationCases()

	for _, tc := range validationCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.ExchangeRate.Validate()

			if tc.ShouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedErr)
			}
		})
	}
}

func TestExchangeRateDateRange(t *testing.T) {
	dateRangeCases := fixtures.DateRangeTestCases()

	for _, tc := range dateRangeCases {
		t.Run(tc.Name, func(t *testing.T) {
			exchangeRate := fixtures.ExchangeRateWithDate(tc.EffectiveDate)

			result := exchangeRate.IsWithinDateRange(tc.TransactionDate)
			assert.Equal(t, tc.ShouldBeValid, result)
		})
	}
}

func TestExchangeRateConvertAmount(t *testing.T) {
	conversionCases := fixtures.ConversionTestCases()

	for _, tc := range conversionCases {
		t.Run(tc.Name, func(t *testing.T) {
			exchangeRate := fixtures.ExchangeRateWithRate(tc.Rate)

			result := exchangeRate.ConvertAmount(tc.Amount)

			// Allow small difference due to floating point precision
			assert.InDelta(t, tc.Expected.Dollars(), result.Dollars(), 0.01)
		})
	}
}

func TestNewExchangeRate(t *testing.T) {
	t.Run("Valid exchange rate creation", func(t *testing.T) {
		effectiveDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

		exchangeRate, err := entities.NewExchangeRate(entities.USD, entities.BRL, 5.20, effectiveDate)

		assert.NoError(t, err)
		assert.NotNil(t, exchangeRate)
		assert.NotEmpty(t, exchangeRate.ID)
		assert.Equal(t, entities.USD, exchangeRate.FromCurrency)
		assert.Equal(t, entities.BRL, exchangeRate.ToCurrency)
		assert.Equal(t, 5.20, exchangeRate.Rate)
		assert.Equal(t, effectiveDate, exchangeRate.EffectiveDate)
		assert.False(t, exchangeRate.RecordDate.IsZero())
	})

	t.Run("Invalid exchange rate creation", func(t *testing.T) {
		// Try to create with invalid rate
		exchangeRate, err := entities.NewExchangeRate(entities.USD, entities.EUR, -1.0, time.Now())

		assert.Error(t, err)
		assert.Nil(t, exchangeRate)
		assert.Contains(t, err.Error(), "must be positive")
	})
}

func TestNewConvertedTransaction(t *testing.T) {
	t.Run("Valid converted transaction", func(t *testing.T) {
		// Setup transaction and exchange rate with compatible dates
		transaction := fixtures.TransactionWithDate(time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC))
		exchangeRate := fixtures.ExchangeRateWithDate(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC))
		exchangeRate.FromCurrency = entities.USD
		exchangeRate.ToCurrency = entities.BRL

		convertedTx, err := entities.NewConvertedTransaction(transaction, entities.BRL, &exchangeRate)

		assert.NoError(t, err)
		assert.NotNil(t, convertedTx)
		assert.Equal(t, transaction, convertedTx.Transaction)
		assert.Equal(t, entities.BRL, convertedTx.TargetCurrency)
		assert.Equal(t, exchangeRate.Rate, convertedTx.ExchangeRate)
		assert.Equal(t, exchangeRate.EffectiveDate, convertedTx.EffectiveDate)

		// Verify converted amount
		expectedAmount := exchangeRate.ConvertAmount(transaction.Amount)
		assert.Equal(t, expectedAmount, convertedTx.ConvertedAmount)
	})

	t.Run("Exchange rate too old", func(t *testing.T) {
		// Transaction in Jan 2024, exchange rate from Jun 2023 (7 months ago)
		transaction := fixtures.TransactionWithDate(time.Date(2024, 1, 20, 0, 0, 0, 0, time.UTC))
		exchangeRate := fixtures.ExchangeRateWithDate(time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC))

		convertedTx, err := entities.NewConvertedTransaction(transaction, entities.BRL, &exchangeRate)

		assert.Error(t, err)
		assert.Nil(t, convertedTx)
		assert.Contains(t, err.Error(), "not within 6 months")
	})

	t.Run("Exchange rate from non-USD", func(t *testing.T) {
		transaction := fixtures.ValidTransaction()
		exchangeRate := fixtures.ExchangeRateWithCurrencies(entities.EUR, entities.BRL)

		convertedTx, err := entities.NewConvertedTransaction(transaction, entities.BRL, &exchangeRate)

		assert.Error(t, err)
		assert.Nil(t, convertedTx)
		assert.Contains(t, err.Error(), "conversion must be from USD")
	})

	t.Run("Exchange rate currency mismatch", func(t *testing.T) {
		transaction := fixtures.ValidTransaction()
		exchangeRate := fixtures.ExchangeRateWithCurrencies(entities.USD, entities.EUR)

		// Try to convert to BRL but exchange rate is for EUR
		convertedTx, err := entities.NewConvertedTransaction(transaction, entities.BRL, &exchangeRate)

		assert.Error(t, err)
		assert.Nil(t, convertedTx)
		assert.Contains(t, err.Error(), "does not match target currency")
	})
}

func TestConvertedTransactionWithFixtures(t *testing.T) {
	t.Run("Valid converted transaction from fixture", func(t *testing.T) {
		convertedTx := fixtures.ValidConvertedTransaction()

		assert.NotEmpty(t, convertedTx.Transaction.ID)
		assert.Equal(t, entities.BRL, convertedTx.TargetCurrency)
		assert.True(t, convertedTx.ExchangeRate > 0)
		assert.True(t, convertedTx.ConvertedAmount > 0)
		assert.False(t, convertedTx.EffectiveDate.IsZero())
	})
}

func TestCurrencyCodeConstants(t *testing.T) {
	// Test that all currency constants are valid
	currencies := []entities.CurrencyCode{
		entities.USD, entities.EUR, entities.BRL,
		entities.GBP, entities.JPY, entities.CAD,
		entities.AUD, entities.CNY,
	}

	for _, currency := range currencies {
		t.Run("Constant_"+currency.String(), func(t *testing.T) {
			assert.True(t, currency.IsValid())
			assert.Len(t, currency.String(), 3)
		})
	}
}
