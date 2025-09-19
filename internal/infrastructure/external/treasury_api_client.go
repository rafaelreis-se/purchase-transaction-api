package external

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/config"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/services"
)

// TreasuryAPIClient implements TreasuryService interface using the real Treasury API
type TreasuryAPIClient struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// TreasuryAPIResponse represents the response structure from Treasury API
type TreasuryAPIResponse struct {
	Data []TreasuryRecord `json:"data"`
	Meta struct {
		Count      int `json:"count"`
		TotalCount int `json:"total-count"`
	} `json:"meta"`
}

// TreasuryRecord represents a single exchange rate record from Treasury API
type TreasuryRecord struct {
	RecordDate          string `json:"record_date"`
	Country             string `json:"country"`
	Currency            string `json:"currency"`
	CountryCurrencyDesc string `json:"country_currency_desc"`
	ExchangeRate        string `json:"exchange_rate"`
	EffectiveDate       string `json:"effective_date"`
}

// NewTreasuryAPIClient creates a new Treasury API client with configuration
func NewTreasuryAPIClient(cfg *config.TreasuryConfig) services.TreasuryService {
	return &TreasuryAPIClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
		timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
	}
}

// FetchExchangeRate retrieves exchange rate from Treasury API for a specific date
func (c *TreasuryAPIClient) FetchExchangeRate(from, to entities.CurrencyCode, date time.Time) (*entities.ExchangeRate, error) {
	startTime := time.Now()

	// Treasury API only supports USD as base currency
	if from != entities.USD {
		slog.Warn("Treasury API only supports USD as base currency",
			"from_currency", string(from),
			"to_currency", string(to),
		)
		return nil, fmt.Errorf("Treasury API only supports USD as base currency, got %s", from)
	}

	// Calculate date range (6 months before the transaction date)
	sixMonthsAgo := date.AddDate(0, -6, 0)

	// Build API URL with filters
	url := c.buildURL(to, sixMonthsAgo, date)

	slog.Info("Calling Treasury API",
		"from_currency", string(from),
		"to_currency", string(to),
		"date", date.Format("2006-01-02"),
		"url", url,
		"currency_filter", c.mapCurrencyCodeToFilter(to),
	)

	// Make HTTP request
	resp, err := c.httpClient.Get(url)
	duration := time.Since(startTime)

	if err != nil {
		slog.Error("Failed to fetch from Treasury API",
			"error", err.Error(),
			"duration", duration,
			"url", url,
		)
		return nil, fmt.Errorf("failed to fetch from Treasury API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Treasury API returned non-200 status",
			"status_code", resp.StatusCode,
			"duration", duration,
			"url", url,
		)
		return nil, fmt.Errorf("Treasury API returned status %d", resp.StatusCode)
	}

	slog.Info("Treasury API call successful",
		"status_code", resp.StatusCode,
		"duration", duration,
	)

	// Parse response
	var apiResponse TreasuryAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		slog.Error("Failed to parse Treasury API response",
			"error", err.Error(),
			"duration", duration,
		)
		return nil, fmt.Errorf("failed to parse Treasury API response: %w", err)
	}

	// Find the most recent rate within the date range
	exchangeRate, err := c.parseExchangeRate(apiResponse.Data, from, to, date)
	if err != nil {
		return nil, err
	}

	return exchangeRate, nil
}

// buildURL constructs the Treasury API URL with appropriate filters
func (c *TreasuryAPIClient) buildURL(currency entities.CurrencyCode, startDate, endDate time.Time) string {
	// Treasury API expects currency in full name format via country_currency_desc
	currencyFilter := c.mapCurrencyCodeToFilter(currency)

	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Build URL with proper encoding
	baseURL := c.baseURL
	params := url.Values{}
	params.Add("fields", "country_currency_desc,exchange_rate,record_date")
	params.Add("filter", fmt.Sprintf("country_currency_desc:eq:%s,record_date:gte:%s,record_date:lte:%s", currencyFilter, startDateStr, endDateStr))
	params.Add("sort", "-record_date")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// mapCurrencyCodeToFilter maps currency codes to Treasury API filter format
func (c *TreasuryAPIClient) mapCurrencyCodeToFilter(code entities.CurrencyCode) string {
	currencyMap := map[entities.CurrencyCode]string{
		entities.EUR: "Euro Zone-Euro",
		entities.GBP: "United Kingdom-Pound", // Back to original from PDF
		entities.JPY: "Japan-Yen",
		entities.CAD: "Canada-Dollar",
		entities.AUD: "Australia-Dollar",
		entities.CNY: "China-Renminbi",
		entities.BRL: "Brazil-Real",
		// Add more mappings as needed
	}

	if filter, exists := currencyMap[code]; exists {
		return filter
	}

	// Fallback to currency code itself
	return string(code)
}

// parseExchangeRate finds the most recent valid exchange rate from API response
func (c *TreasuryAPIClient) parseExchangeRate(records []TreasuryRecord, from, to entities.CurrencyCode, transactionDate time.Time) (*entities.ExchangeRate, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no exchange rate found for %s within 6 months of %s", to, transactionDate.Format("2006-01-02"))
	}

	// Records are sorted by record_date descending, so take the first valid one
	for _, record := range records {
		rate, err := c.parseRecord(record, from, to)
		if err != nil {
			continue // Skip invalid records
		}

		// Verify the rate is within the 6-month rule
		if rate.IsWithinDateRange(transactionDate) {
			return rate, nil
		}
	}

	return nil, fmt.Errorf("no suitable exchange rate found for %s within 6 months of %s", to, transactionDate.Format("2006-01-02"))
}

// parseRecord converts a Treasury API record to an ExchangeRate entity
func (c *TreasuryAPIClient) parseRecord(record TreasuryRecord, from, to entities.CurrencyCode) (*entities.ExchangeRate, error) {
	// Parse exchange rate
	var rate float64
	if _, err := fmt.Sscanf(record.ExchangeRate, "%f", &rate); err != nil {
		return nil, fmt.Errorf("invalid exchange rate format: %s", record.ExchangeRate)
	}

	// Parse record date (Treasury API only has record_date, not effective_date)
	recordDate, err := time.Parse("2006-01-02", record.RecordDate)
	if err != nil {
		return nil, fmt.Errorf("invalid record date format: %s", record.RecordDate)
	}

	// Create ExchangeRate entity - use record_date as both effective and record date
	exchangeRate := &entities.ExchangeRate{
		ID:            uuid.New(),
		FromCurrency:  from,
		ToCurrency:    to,
		Rate:          rate,
		EffectiveDate: recordDate, // Use record_date as effective_date
		RecordDate:    recordDate,
		CreatedAt:     time.Now(),
	}

	return exchangeRate, nil
}
