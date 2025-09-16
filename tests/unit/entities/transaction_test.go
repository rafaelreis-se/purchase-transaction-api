package entities_test

import (
	"fmt"
	"testing"

	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/tests/fixtures"
	"github.com/stretchr/testify/assert"
)

func TestNewMoneyWithFixtures(t *testing.T) {
	testCases := fixtures.MoneyTestCases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			result := entities.NewMoney(tc.Dollars)
			assert.Equal(t, tc.Expected, result)
		})
	}
}

func TestMoneyDollars(t *testing.T) {
	testCases := fixtures.MoneyMethodTestCases().DollarsTests

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Money.Dollars()
			assert.Equal(t, tt.Expected, result)
		})
	}
}

func TestMoneyCents(t *testing.T) {
	testCases := fixtures.MoneyMethodTestCases().CentsTests

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Money.Cents()
			assert.Equal(t, tt.Expected, result)
		})
	}
}

func TestMoneyIsPositive(t *testing.T) {
	testCases := fixtures.MoneyMethodTestCases().IsPositiveTests

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			result := tt.Money.IsPositive()
			assert.Equal(t, tt.Expected, result)
		})
	}
}

func TestTransactionValidationWithFixtures(t *testing.T) {
	// Test valid cases
	t.Run("Valid transaction", func(t *testing.T) {
		tx := fixtures.ValidTransaction()
		err := tx.Validate()
		assert.NoError(t, err)
	})

	// Test invalid cases using fixtures
	invalidCases := fixtures.InvalidTransactions()
	for _, tc := range invalidCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Transaction.Validate()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.ExpectedErr)
		})
	}
}

func TestTransactionCreationWithFixtures(t *testing.T) {
	t.Run("Default transaction", func(t *testing.T) {
		tx := fixtures.ValidTransaction()

		assert.NotEmpty(t, tx.ID)
		assert.Equal(t, "Test Purchase", tx.Description)
		assert.True(t, tx.Amount > 0)
	})

	t.Run("Custom description", func(t *testing.T) {
		tx := fixtures.TransactionWithDescription("Custom desc")

		assert.Equal(t, "Custom desc", tx.Description)
	})

	t.Run("Custom amount", func(t *testing.T) {
		tx := fixtures.TransactionWithAmount(50.75)

		assert.Equal(t, entities.NewMoney(50.75), tx.Amount)
		assert.Equal(t, 50.75, tx.Amount.Dollars())
	})
}

func TestMinimalTransaction(t *testing.T) {
	tx := fixtures.MinimalTransaction()

	assert.NotEmpty(t, tx.ID)
	assert.Equal(t, "Test", tx.Description)
	assert.Equal(t, entities.NewMoney(1.00), tx.Amount)
}

func TestTransactionValidationEdgeCases(t *testing.T) {
	edgeCases := fixtures.ValidationEdgeCases()

	for _, tc := range edgeCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Transaction.Validate()

			if tc.ShouldPass {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.ExpectedErr)
			}
		})
	}
}

func TestMoneyRoundTrip(t *testing.T) {
	testCases := fixtures.RoundTripTestCases()

	for _, dollars := range testCases {
		t.Run(fmt.Sprintf("%.2f", dollars), func(t *testing.T) {
			money := entities.NewMoney(dollars)
			result := money.Dollars()

			// Should be equal within 0.01 (1 cent) due to rounding
			assert.InDelta(t, dollars, result, 0.01)
		})
	}
}
