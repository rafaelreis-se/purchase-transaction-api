package fixtures

import (
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// ValidTransaction creates a transaction with valid default values
func ValidTransaction() entities.Transaction {
	return entities.Transaction{
		ID:          uuid.New(),
		Description: "Test Purchase",
		Date:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Amount:      entities.NewMoney(99.99),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TransactionWithDescription creates a transaction with custom description
func TransactionWithDescription(description string) entities.Transaction {
	tx := ValidTransaction()
	tx.Description = description
	return tx
}

// TransactionWithAmount creates a transaction with custom amount
func TransactionWithAmount(dollars float64) entities.Transaction {
	tx := ValidTransaction()
	tx.Amount = entities.NewMoney(dollars)
	return tx
}

// TransactionWithDate creates a transaction with custom date
func TransactionWithDate(date time.Time) entities.Transaction {
	tx := ValidTransaction()
	tx.Date = date
	return tx
}

// TransactionWithID creates a transaction with specific UUID (useful for deterministic tests)
func TransactionWithID(id uuid.UUID) entities.Transaction {
	tx := ValidTransaction()
	tx.ID = id
	return tx
}

// MinimalTransaction creates a transaction with minimal required fields
func MinimalTransaction() entities.Transaction {
	return entities.Transaction{
		ID:          uuid.New(),
		Description: "Test",
		Date:        time.Now(),
		Amount:      entities.NewMoney(1.00),
	}
}

// InvalidTransactions returns test cases for validation testing
func InvalidTransactions() []struct {
	Name        string
	Transaction entities.Transaction
	ExpectedErr string
} {
	return []struct {
		Name        string
		Transaction entities.Transaction
		ExpectedErr string
	}{
		{
			Name: "Empty description",
			Transaction: entities.Transaction{
				ID:          uuid.New(),
				Description: "",
				Date:        time.Now(),
				Amount:      entities.NewMoney(10.00),
			},
			ExpectedErr: "description is required",
		},
		{
			Name: "Description too long",
			Transaction: entities.Transaction{
				ID:          uuid.New(),
				Description: "This description is way too long and exceeds the fifty character limit that we have set",
				Date:        time.Now(),
				Amount:      entities.NewMoney(10.00),
			},
			ExpectedErr: "description must not exceed 50 characters",
		},
		{
			Name: "Zero date",
			Transaction: entities.Transaction{
				ID:          uuid.New(),
				Description: "Valid description",
				Date:        time.Time{}, // Zero time
				Amount:      entities.NewMoney(10.00),
			},
			ExpectedErr: "transaction date is required",
		},
		{
			Name: "Zero amount",
			Transaction: entities.Transaction{
				ID:          uuid.New(),
				Description: "Valid description",
				Date:        time.Now(),
				Amount:      entities.Money(0),
			},
			ExpectedErr: "purchase amount must be positive",
		},
		{
			Name: "Negative amount",
			Transaction: entities.Transaction{
				ID:          uuid.New(),
				Description: "Valid description",
				Date:        time.Now(),
				Amount:      entities.Money(-100),
			},
			ExpectedErr: "purchase amount must be positive",
		},
	}
}

// MoneyTestCases returns test cases for Money type testing
func MoneyTestCases() []struct {
	Name     string
	Dollars  float64
	Expected entities.Money
} {
	return []struct {
		Name     string
		Dollars  float64
		Expected entities.Money
	}{
		{"Whole dollars", 10.00, entities.Money(1000)},
		{"With cents", 19.99, entities.Money(1999)},
		{"Zero", 0.00, entities.Money(0)},
		{"One cent", 0.01, entities.Money(1)},
		{"Large amount", 1234.56, entities.Money(123456)},
		{"Rounding down", 19.994, entities.Money(1999)},
		{"Rounding up", 19.996, entities.Money(2000)},
	}
}

// MoneyMethodTestCases returns test cases for Money methods (Dollars, Cents, IsPositive)
func MoneyMethodTestCases() struct {
	DollarsTests []struct {
		Name     string
		Money    entities.Money
		Expected float64
	}
	CentsTests []struct {
		Name     string
		Money    entities.Money
		Expected int64
	}
	IsPositiveTests []struct {
		Name     string
		Money    entities.Money
		Expected bool
	}
} {
	return struct {
		DollarsTests []struct {
			Name     string
			Money    entities.Money
			Expected float64
		}
		CentsTests []struct {
			Name     string
			Money    entities.Money
			Expected int64
		}
		IsPositiveTests []struct {
			Name     string
			Money    entities.Money
			Expected bool
		}
	}{
		DollarsTests: []struct {
			Name     string
			Money    entities.Money
			Expected float64
		}{
			{"Whole dollars", entities.Money(1000), 10.00},
			{"With cents", entities.Money(1999), 19.99},
			{"Zero", entities.Money(0), 0.00},
			{"One cent", entities.Money(1), 0.01},
			{"Large amount", entities.Money(123456), 1234.56},
		},
		CentsTests: []struct {
			Name     string
			Money    entities.Money
			Expected int64
		}{
			{"Whole dollars", entities.Money(1000), 1000},
			{"With cents", entities.Money(1999), 1999},
			{"Zero", entities.Money(0), 0},
			{"One cent", entities.Money(1), 1},
			{"Negative", entities.Money(-500), -500},
		},
		IsPositiveTests: []struct {
			Name     string
			Money    entities.Money
			Expected bool
		}{
			{"Positive amount", entities.Money(1000), true},
			{"Small positive", entities.Money(1), true},
			{"Zero", entities.Money(0), false},
			{"Negative", entities.Money(-100), false},
		},
	}
}

// ValidationEdgeCases returns edge case scenarios for validation testing
func ValidationEdgeCases() []struct {
	Name        string
	Transaction entities.Transaction
	ShouldPass  bool
	ExpectedErr string
} {
	return []struct {
		Name        string
		Transaction entities.Transaction
		ShouldPass  bool
		ExpectedErr string
	}{
		{
			Name:        "Description exactly 50 characters",
			Transaction: TransactionWithDescription("This description has exactly fifty characters!!"),
			ShouldPass:  true,
		},
		{
			Name:        "Description with 51 characters",
			Transaction: TransactionWithDescription("This description has exactly fifty-one characters!!"),
			ShouldPass:  false,
			ExpectedErr: "must not exceed 50 characters",
		},
		{
			Name:        "Minimal valid amount",
			Transaction: TransactionWithAmount(0.01),
			ShouldPass:  true,
		},
	}
}

// RoundTripTestCases returns test cases for money round-trip conversion
func RoundTripTestCases() []float64 {
	return []float64{10.00, 19.99, 0.01, 1234.56, 99.95, 0.99, 1000.00}
}
