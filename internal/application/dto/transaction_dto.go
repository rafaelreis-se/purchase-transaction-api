package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// CreateTransactionRequest represents the input for creating a new transaction
type CreateTransactionRequest struct {
	Description string    `json:"description" validate:"required,max=50"`
	Date        time.Time `json:"date" validate:"required"`
	Amount      float64   `json:"amount" validate:"required,gt=0"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetTransactionResponse represents the response for retrieving a transaction
type GetTransactionResponse struct {
	ID          uuid.UUID `json:"id"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListTransactionsRequest represents the input for listing transactions with pagination
type ListTransactionsRequest struct {
	Page int `json:"page" validate:"min=1" default:"1"`
	Size int `json:"size" validate:"min=1,max=100" default:"20"`
}

// ListTransactionsResponse represents the response for listing transactions
type ListTransactionsResponse struct {
	Data       []GetTransactionResponse `json:"data"`
	Page       int                      `json:"page"`
	Size       int                      `json:"size"`
	Total      int64                    `json:"total"`
	TotalPages int                      `json:"total_pages"`
}

// ConvertTransactionRequest represents the input for currency conversion
type ConvertTransactionRequest struct {
	TransactionID  uuid.UUID             `json:"transaction_id" validate:"required"`
	TargetCurrency entities.CurrencyCode `json:"target_currency" validate:"required"`
}

// ConvertTransactionResponse represents the response after currency conversion
type ConvertTransactionResponse struct {
	Transaction     GetTransactionResponse `json:"transaction"`
	TargetCurrency  entities.CurrencyCode  `json:"target_currency"`
	ExchangeRate    float64                `json:"exchange_rate"`
	ConvertedAmount float64                `json:"converted_amount"`
	EffectiveDate   time.Time              `json:"effective_date"`
}

// ToEntity converts CreateTransactionRequest to Transaction entity
func (req *CreateTransactionRequest) ToEntity() *entities.Transaction {
	return &entities.Transaction{
		ID:          uuid.New(),
		Description: req.Description,
		Date:        req.Date,
		Amount:      entities.NewMoney(req.Amount),
		CreatedAt:   time.Now(),
	}
}

// FromEntity converts Transaction entity to CreateTransactionResponse
func NewCreateTransactionResponse(transaction *entities.Transaction) *CreateTransactionResponse {
	return &CreateTransactionResponse{
		ID:          transaction.ID,
		Description: transaction.Description,
		Date:        transaction.Date,
		Amount:      transaction.Amount.Dollars(),
		CreatedAt:   transaction.CreatedAt,
	}
}

// FromEntity converts Transaction entity to GetTransactionResponse
func NewGetTransactionResponse(transaction *entities.Transaction) *GetTransactionResponse {
	return &GetTransactionResponse{
		ID:          transaction.ID,
		Description: transaction.Description,
		Date:        transaction.Date,
		Amount:      transaction.Amount.Dollars(),
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
	}
}

// NewListTransactionsResponse creates a paginated response for listing transactions
func NewListTransactionsResponse(transactions []entities.Transaction, page, size int, total int64) *ListTransactionsResponse {
	responses := make([]GetTransactionResponse, len(transactions))
	for i, tx := range transactions {
		responses[i] = *NewGetTransactionResponse(&tx)
	}

	totalPages := int((total + int64(size) - 1) / int64(size)) // Ceiling division

	return &ListTransactionsResponse{
		Data:       responses,
		Page:       page,
		Size:       size,
		Total:      total,
		TotalPages: totalPages,
	}
}

// NewConvertTransactionResponse converts ConvertedTransaction entity to response
func NewConvertTransactionResponse(convertedTx *entities.ConvertedTransaction) *ConvertTransactionResponse {
	return &ConvertTransactionResponse{
		Transaction:     *NewGetTransactionResponse(&convertedTx.Transaction),
		TargetCurrency:  convertedTx.TargetCurrency,
		ExchangeRate:    convertedTx.ExchangeRate,
		ConvertedAmount: convertedTx.ConvertedAmount.Dollars(),
		EffectiveDate:   convertedTx.EffectiveDate,
	}
}
