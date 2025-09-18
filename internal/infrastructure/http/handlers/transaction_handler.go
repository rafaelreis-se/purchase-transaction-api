package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/dto"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
)

// TransactionHandler handles HTTP requests for transaction operations
type TransactionHandler struct {
	createTransactionUseCase  *usecases.CreateTransactionUseCase
	getTransactionUseCase     *usecases.GetTransactionUseCase
	listTransactionsUseCase   *usecases.ListTransactionsUseCase
	convertTransactionUseCase *usecases.ConvertTransactionUseCase
}

// NewTransactionHandler creates a new TransactionHandler
func NewTransactionHandler(
	createTransactionUseCase *usecases.CreateTransactionUseCase,
	getTransactionUseCase *usecases.GetTransactionUseCase,
	listTransactionsUseCase *usecases.ListTransactionsUseCase,
	convertTransactionUseCase *usecases.ConvertTransactionUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		createTransactionUseCase:  createTransactionUseCase,
		getTransactionUseCase:     getTransactionUseCase,
		listTransactionsUseCase:   listTransactionsUseCase,
		convertTransactionUseCase: convertTransactionUseCase,
	}
}

// CreateTransaction handles POST /transactions
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var request dto.CreateTransactionRequest

	// Bind JSON request to DTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Execute use case
	response, err := h.createTransactionUseCase.Execute(&request)
	if err != nil {
		// Check error type for appropriate status code
		statusCode := http.StatusInternalServerError
		if isValidationError(err) {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to create transaction",
			"details": err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusCreated, response)
}

// GetTransaction handles GET /transactions/:id
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	// Parse UUID from path parameter
	idParam := c.Param("id")
	transactionID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid transaction ID format",
			"details": "Transaction ID must be a valid UUID",
		})
		return
	}

	// Execute use case
	response, err := h.getTransactionUseCase.Execute(transactionID)
	if err != nil {
		// Check if transaction not found
		statusCode := http.StatusInternalServerError
		if isNotFoundError(err) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to retrieve transaction",
			"details": err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, response)
}

// ListTransactions handles GET /transactions
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	// Parse query parameters with defaults
	page := 1
	size := 20

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if sizeParam := c.Query("size"); sizeParam != "" {
		if s, err := strconv.Atoi(sizeParam); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	// Create request DTO
	request := &dto.ListTransactionsRequest{
		Page: page,
		Size: size,
	}

	// Execute use case
	response, err := h.listTransactionsUseCase.Execute(request)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if isValidationError(err) {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to retrieve transactions",
			"details": err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, response)
}

// ConvertTransaction handles POST /transactions/:id/convert
func (h *TransactionHandler) ConvertTransaction(c *gin.Context) {
	// Parse UUID from path parameter
	idParam := c.Param("id")
	transactionID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid transaction ID format",
			"details": "Transaction ID must be a valid UUID",
		})
		return
	}

	// Parse request body for target currency
	var requestBody struct {
		TargetCurrency string `json:"target_currency" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Create use case request
	request := &dto.ConvertTransactionRequest{
		TransactionID:  transactionID,
		TargetCurrency: entities.CurrencyCode(requestBody.TargetCurrency),
	}

	// Execute use case
	response, err := h.convertTransactionUseCase.Execute(request)
	if err != nil {
		// Determine appropriate status code
		statusCode := http.StatusInternalServerError
		if isValidationError(err) {
			statusCode = http.StatusBadRequest
		} else if isNotFoundError(err) {
			statusCode = http.StatusNotFound
		} else if isExchangeRateNotFoundError(err) {
			statusCode = http.StatusUnprocessableEntity
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to convert transaction",
			"details": err.Error(),
		})
		return
	}

	// Return successful response
	c.JSON(http.StatusOK, response)
}

// Helper functions for error classification

func isValidationError(err error) bool {
	return contains(err.Error(), "validation failed") ||
		contains(err.Error(), "invalid") ||
		contains(err.Error(), "required")
}

func isNotFoundError(err error) bool {
	return contains(err.Error(), "not found")
}

func isExchangeRateNotFoundError(err error) bool {
	return contains(err.Error(), "no suitable exchange rate found") ||
		contains(err.Error(), "within 6 months")
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func indexOf(s, substr string) int {
	return strings.Index(s, substr)
}
