package services

import (
	"time"

	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// TreasuryService defines the contract for fetching exchange rates from Treasury API
type TreasuryService interface {
	// FetchExchangeRate retrieves exchange rate from Treasury API for a specific date
	// Returns the most recent rate within 6 months before the given date
	FetchExchangeRate(from, to entities.CurrencyCode, date time.Time) (*entities.ExchangeRate, error)
}
