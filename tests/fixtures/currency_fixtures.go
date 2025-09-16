package fixtures

import (
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// ValidExchangeRate creates an exchange rate with valid default values
func ValidExchangeRate() entities.ExchangeRate {
	return entities.ExchangeRate{
		ID:            uuid.New(),
		FromCurrency:  entities.USD,
		ToCurrency:    entities.BRL,
		Rate:          5.20,
		EffectiveDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		RecordDate:    time.Now(),
	}
}

// ExchangeRateWithRate creates an exchange rate with custom rate
func ExchangeRateWithRate(rate float64) entities.ExchangeRate {
	er := ValidExchangeRate()
	er.Rate = rate
	return er
}

// ExchangeRateWithCurrencies creates an exchange rate with custom currencies
func ExchangeRateWithCurrencies(from, to entities.CurrencyCode) entities.ExchangeRate {
	er := ValidExchangeRate()
	er.FromCurrency = from
	er.ToCurrency = to
	return er
}

// ExchangeRateWithDate creates an exchange rate with custom effective date
func ExchangeRateWithDate(effectiveDate time.Time) entities.ExchangeRate {
	er := ValidExchangeRate()
	er.EffectiveDate = effectiveDate
	return er
}

// ValidConvertedTransaction creates a converted transaction with valid data
func ValidConvertedTransaction() entities.ConvertedTransaction {
	return entities.ConvertedTransaction{
		Transaction:     ValidTransaction(),
		TargetCurrency:  entities.BRL,
		ExchangeRate:    5.20,
		ConvertedAmount: entities.NewMoney(519.48), // 99.99 * 5.20
		EffectiveDate:   time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

// CurrencyCodeTestCases returns test cases for CurrencyCode validation
func CurrencyCodeTestCases() struct {
	ValidCodes []struct {
		Input    string
		Expected entities.CurrencyCode
	}
	InvalidCodes []struct {
		Input       string
		ExpectedErr string
	}
} {
	return struct {
		ValidCodes []struct {
			Input    string
			Expected entities.CurrencyCode
		}
		InvalidCodes []struct {
			Input       string
			ExpectedErr string
		}
	}{
		ValidCodes: []struct {
			Input    string
			Expected entities.CurrencyCode
		}{
			{"USD", entities.USD},
			{"usd", entities.USD},   // lowercase should be normalized
			{" EUR ", entities.EUR}, // with spaces
			{"BRL", entities.BRL},
			{"GBP", entities.GBP},
		},
		InvalidCodes: []struct {
			Input       string
			ExpectedErr string
		}{
			{"", "must be exactly 3 characters"},
			{"US", "must be exactly 3 characters"},
			{"USDD", "must be exactly 3 characters"},
			{"123", "invalid currency code format"},
			{"U$D", "invalid currency code format"},
		},
	}
}

// ExchangeRateValidationCases returns test cases for ExchangeRate validation
func ExchangeRateValidationCases() []struct {
	Name         string
	ExchangeRate entities.ExchangeRate
	ShouldPass   bool
	ExpectedErr  string
} {
	return []struct {
		Name         string
		ExchangeRate entities.ExchangeRate
		ShouldPass   bool
		ExpectedErr  string
	}{
		{
			Name:         "Valid exchange rate",
			ExchangeRate: ValidExchangeRate(),
			ShouldPass:   true,
		},
		{
			Name:         "Invalid from currency",
			ExchangeRate: ExchangeRateWithCurrencies("XX", entities.BRL),
			ShouldPass:   false,
			ExpectedErr:  "invalid from_currency",
		},
		{
			Name:         "Invalid to currency",
			ExchangeRate: ExchangeRateWithCurrencies(entities.USD, "YY"),
			ShouldPass:   false,
			ExpectedErr:  "invalid to_currency",
		},
		{
			Name:         "Same currencies",
			ExchangeRate: ExchangeRateWithCurrencies(entities.USD, entities.USD),
			ShouldPass:   false,
			ExpectedErr:  "cannot be the same",
		},
		{
			Name:         "Zero rate",
			ExchangeRate: ExchangeRateWithRate(0.0),
			ShouldPass:   false,
			ExpectedErr:  "must be positive",
		},
		{
			Name:         "Negative rate",
			ExchangeRate: ExchangeRateWithRate(-1.5),
			ShouldPass:   false,
			ExpectedErr:  "must be positive",
		},
		{
			Name:         "Zero effective date",
			ExchangeRate: ExchangeRateWithDate(time.Time{}),
			ShouldPass:   false,
			ExpectedErr:  "effective_date is required",
		},
	}
}

// DateRangeTestCases returns test cases for date range validation
func DateRangeTestCases() []struct {
	Name            string
	TransactionDate time.Time
	EffectiveDate   time.Time
	ShouldBeValid   bool
} {
	transactionDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	return []struct {
		Name            string
		TransactionDate time.Time
		EffectiveDate   time.Time
		ShouldBeValid   bool
	}{
		{
			Name:            "Same date",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate,
			ShouldBeValid:   true,
		},
		{
			Name:            "One day before",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate.AddDate(0, 0, -1),
			ShouldBeValid:   true,
		},
		{
			Name:            "Exactly 6 months before",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate.AddDate(0, -6, 0),
			ShouldBeValid:   true,
		},
		{
			Name:            "5 months before",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate.AddDate(0, -5, 0),
			ShouldBeValid:   true,
		},
		{
			Name:            "7 months before (too old)",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate.AddDate(0, -7, 0),
			ShouldBeValid:   false,
		},
		{
			Name:            "One day after (future)",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate.AddDate(0, 0, 1),
			ShouldBeValid:   false,
		},
		{
			Name:            "One month after (future)",
			TransactionDate: transactionDate,
			EffectiveDate:   transactionDate.AddDate(0, 1, 0),
			ShouldBeValid:   false,
		},
	}
}

// ConversionTestCases returns test cases for amount conversion
func ConversionTestCases() []struct {
	Name     string
	Amount   entities.Money
	Rate     float64
	Expected entities.Money
} {
	return []struct {
		Name     string
		Amount   entities.Money
		Rate     float64
		Expected entities.Money
	}{
		{
			Name:     "USD to BRL",
			Amount:   entities.NewMoney(100.00),
			Rate:     5.20,
			Expected: entities.NewMoney(520.00),
		},
		{
			Name:     "USD to EUR",
			Amount:   entities.NewMoney(50.00),
			Rate:     0.85,
			Expected: entities.NewMoney(42.50),
		},
		{
			Name:     "Small amount",
			Amount:   entities.NewMoney(1.00),
			Rate:     1.25,
			Expected: entities.NewMoney(1.25),
		},
		{
			Name:     "High precision rate",
			Amount:   entities.NewMoney(99.99),
			Rate:     5.196743,
			Expected: entities.NewMoney(519.62), // Rounded to nearest cent
		},
	}
}
