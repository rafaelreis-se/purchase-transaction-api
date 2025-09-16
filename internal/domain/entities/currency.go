package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CurrencyCode represents a 3-letter ISO currency code
type CurrencyCode string

// Standard currency codes
const (
	USD CurrencyCode = "USD"
	EUR CurrencyCode = "EUR"
	BRL CurrencyCode = "BRL"
	GBP CurrencyCode = "GBP"
	JPY CurrencyCode = "JPY"
	CAD CurrencyCode = "CAD"
	AUD CurrencyCode = "AUD"
	CNY CurrencyCode = "CNY"
)

// ExchangeRate represents a currency exchange rate from Treasury API
type ExchangeRate struct {
	ID            uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	FromCurrency  CurrencyCode `json:"from_currency" gorm:"not null"`
	ToCurrency    CurrencyCode `json:"to_currency" gorm:"not null"`
	Rate          float64      `json:"rate" gorm:"not null" validate:"required,gt=0"`
	EffectiveDate time.Time    `json:"effective_date" gorm:"not null" validate:"required"`
	RecordDate    time.Time    `json:"record_date" gorm:"not null"`
	CreatedAt     time.Time    `json:"created_at" gorm:"autoCreateTime"`
}

// ConvertedTransaction represents a transaction with currency conversion applied
type ConvertedTransaction struct {
	Transaction     Transaction  `json:"transaction"`
	TargetCurrency  CurrencyCode `json:"target_currency"`
	ExchangeRate    float64      `json:"exchange_rate"`
	ConvertedAmount Money        `json:"converted_amount"`
	EffectiveDate   time.Time    `json:"effective_date"`
}

// String returns the currency code as string
func (c CurrencyCode) String() string {
	return string(c)
}

// IsValid checks if the currency code is valid (3 letters, uppercase)
func (c CurrencyCode) IsValid() bool {
	code := string(c)
	if len(code) != 3 {
		return false
	}
	
	// Check if all characters are uppercase letters
	for _, char := range code {
		if char < 'A' || char > 'Z' {
			return false
		}
	}
	
	return true
}

// NewCurrencyCode creates a new currency code from string with validation
func NewCurrencyCode(code string) (CurrencyCode, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))

	if len(normalized) != 3 {
		return "", fmt.Errorf("currency code must be exactly 3 characters, got %d", len(normalized))
	}

	currencyCode := CurrencyCode(normalized)
	if !currencyCode.IsValid() {
		return "", fmt.Errorf("invalid currency code format: %s", code)
	}

	return currencyCode, nil
}

// Validate performs business rule validation for ExchangeRate
func (er *ExchangeRate) Validate() error {
	if !er.FromCurrency.IsValid() {
		return fmt.Errorf("invalid from_currency: %s", er.FromCurrency)
	}

	if !er.ToCurrency.IsValid() {
		return fmt.Errorf("invalid to_currency: %s", er.ToCurrency)
	}

	if er.FromCurrency == er.ToCurrency {
		return fmt.Errorf("from_currency and to_currency cannot be the same")
	}

	if er.Rate <= 0 {
		return fmt.Errorf("exchange rate must be positive, got %f", er.Rate)
	}

	if er.EffectiveDate.IsZero() {
		return fmt.Errorf("effective_date is required")
	}

	return nil
}

// IsWithinDateRange checks if the exchange rate is within 6 months before the given date
func (er *ExchangeRate) IsWithinDateRange(transactionDate time.Time) bool {
	sixMonthsAgo := transactionDate.AddDate(0, -6, 0)
	return !er.EffectiveDate.Before(sixMonthsAgo) && !er.EffectiveDate.After(transactionDate)
}

// ConvertAmount converts a Money amount using this exchange rate
func (er *ExchangeRate) ConvertAmount(amount Money) Money {
	dollars := amount.Dollars()
	convertedDollars := dollars * er.Rate
	return NewMoney(convertedDollars)
}

// NewExchangeRate creates a new exchange rate with validation
func NewExchangeRate(from, to CurrencyCode, rate float64, effectiveDate time.Time) (*ExchangeRate, error) {
	exchangeRate := &ExchangeRate{
		ID:            uuid.New(),
		FromCurrency:  from,
		ToCurrency:    to,
		Rate:          rate,
		EffectiveDate: effectiveDate,
		RecordDate:    time.Now(),
	}

	if err := exchangeRate.Validate(); err != nil {
		return nil, err
	}

	return exchangeRate, nil
}

// NewConvertedTransaction creates a converted transaction with proper validation
func NewConvertedTransaction(tx Transaction, targetCurrency CurrencyCode, exchangeRate *ExchangeRate) (*ConvertedTransaction, error) {
	if !exchangeRate.IsWithinDateRange(tx.Date) {
		return nil, fmt.Errorf("exchange rate date %v is not within 6 months of transaction date %v",
			exchangeRate.EffectiveDate, tx.Date)
	}

	if exchangeRate.FromCurrency != USD {
		return nil, fmt.Errorf("conversion must be from USD, got %s", exchangeRate.FromCurrency)
	}

	if exchangeRate.ToCurrency != targetCurrency {
		return nil, fmt.Errorf("exchange rate currency %s does not match target currency %s",
			exchangeRate.ToCurrency, targetCurrency)
	}

	convertedAmount := exchangeRate.ConvertAmount(tx.Amount)

	return &ConvertedTransaction{
		Transaction:     tx,
		TargetCurrency:  targetCurrency,
		ExchangeRate:    exchangeRate.Rate,
		ConvertedAmount: convertedAmount,
		EffectiveDate:   exchangeRate.EffectiveDate,
	}, nil
}
