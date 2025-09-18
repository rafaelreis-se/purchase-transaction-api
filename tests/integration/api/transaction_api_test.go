package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/database"
	httpInfra "github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a test router with real dependencies
func setupTestRouter(t *testing.T) (*gin.Engine, func()) {
	// Create in-memory database
	db, err := database.NewSQLiteDB(":memory:")
	require.NoError(t, err)

	// Initialize repositories
	transactionRepo := database.NewTransactionRepository(db.GetDB())
	exchangeRateRepo := database.NewExchangeRateRepository(db.GetDB())

	// Initialize validator
	validator := validator.New()

	// Initialize use cases
	createTransactionUseCase := usecases.NewCreateTransactionUseCase(transactionRepo, validator)
	getTransactionUseCase := usecases.NewGetTransactionUseCase(transactionRepo)
	listTransactionsUseCase := usecases.NewListTransactionsUseCase(transactionRepo, validator)
	convertTransactionUseCase := usecases.NewConvertTransactionUseCase(transactionRepo, exchangeRateRepo, validator)

	// Initialize handlers
	transactionHandler := handlers.NewTransactionHandler(
		createTransactionUseCase,
		getTransactionUseCase,
		listTransactionsUseCase,
		convertTransactionUseCase,
	)

	// Initialize router
	router := httpInfra.NewRouter(transactionHandler)
	ginRouter := router.SetupRoutes()

	// Cleanup function
	cleanup := func() {
		db.Close()
	}

	return ginRouter, cleanup
}

func TestCreateTransactionAPI(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Successful transaction creation", func(t *testing.T) {
		// Arrange
		requestBody := map[string]interface{}{
			"description": "Test Purchase",
			"date":        "2024-01-15T10:30:00Z",
			"amount":      99.99,
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Act
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.NotEmpty(t, response["id"])
		assert.Equal(t, "Test Purchase", response["description"])
		assert.Equal(t, 99.99, response["amount"])
		assert.NotEmpty(t, response["created_at"])
	})

	t.Run("Invalid request - missing required field", func(t *testing.T) {
		// Arrange
		requestBody := map[string]interface{}{
			"date":   "2024-01-15T10:30:00Z",
			"amount": 99.99,
			// Missing description
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Act
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Failed to create transaction")
	})

	t.Run("Invalid request - negative amount", func(t *testing.T) {
		// Arrange
		requestBody := map[string]interface{}{
			"description": "Test Purchase",
			"date":        "2024-01-15T10:30:00Z",
			"amount":      -10.50,
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Act
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Failed to create transaction")
	})

	t.Run("Invalid JSON format", func(t *testing.T) {
		// Act
		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetTransactionAPI(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Get existing transaction", func(t *testing.T) {
		// Arrange - Create a transaction first
		createReq := map[string]interface{}{
			"description": "Test Purchase",
			"date":        "2024-01-15T10:30:00Z",
			"amount":      99.99,
		}
		jsonBody, _ := json.Marshal(createReq)

		createHttpReq := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createHttpReq)

		require.Equal(t, http.StatusCreated, createW.Code)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)
		transactionID := createResponse["id"].(string)

		// Act - Get the transaction
		getReq := httptest.NewRequest("GET", "/api/v1/transactions/"+transactionID, nil)
		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		// Assert
		assert.Equal(t, http.StatusOK, getW.Code)

		var getResponse map[string]interface{}
		err = json.Unmarshal(getW.Body.Bytes(), &getResponse)
		require.NoError(t, err)

		assert.Equal(t, transactionID, getResponse["id"])
		assert.Equal(t, "Test Purchase", getResponse["description"])
		assert.Equal(t, 99.99, getResponse["amount"])
	})

	t.Run("Get non-existing transaction", func(t *testing.T) {
		// Arrange
		nonExistentID := uuid.New().String()

		// Act
		req := httptest.NewRequest("GET", "/api/v1/transactions/"+nonExistentID, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Failed to retrieve transaction")
		assert.Contains(t, response["details"], "not found")
	})

	t.Run("Invalid UUID format", func(t *testing.T) {
		// Act
		req := httptest.NewRequest("GET", "/api/v1/transactions/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid transaction ID format")
	})
}

func TestListTransactionsAPI(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	// Helper function to create a transaction
	createTransaction := func(description string, amount float64) {
		requestBody := map[string]interface{}{
			"description": description,
			"date":        time.Now().Format(time.RFC3339),
			"amount":      amount,
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	t.Run("List transactions with default pagination", func(t *testing.T) {
		// Arrange - Create some transactions
		createTransaction("Transaction 1", 10.00)
		createTransaction("Transaction 2", 20.00)
		createTransaction("Transaction 3", 30.00)

		// Act
		req := httptest.NewRequest("GET", "/api/v1/transactions", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(20), response["size"])
		assert.Equal(t, float64(3), response["total"])
		assert.Equal(t, float64(1), response["total_pages"])

		data := response["data"].([]interface{})
		assert.Len(t, data, 3)
	})

	t.Run("List transactions with custom pagination", func(t *testing.T) {
		// Act
		req := httptest.NewRequest("GET", "/api/v1/transactions?page=1&size=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(1), response["page"])
		assert.Equal(t, float64(2), response["size"])

		data := response["data"].([]interface{})
		assert.LessOrEqual(t, len(data), 2) // Should have at most 2 items
	})

	t.Run("List transactions - empty result", func(t *testing.T) {
		// Create fresh router with empty database
		freshRouter, cleanup := setupTestRouter(t)
		defer cleanup()

		// Act
		req := httptest.NewRequest("GET", "/api/v1/transactions", nil)
		w := httptest.NewRecorder()
		freshRouter.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, float64(0), response["total"])
		data := response["data"].([]interface{})
		assert.Empty(t, data)
	})
}

func TestConvertTransactionAPI(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Convert transaction - no exchange rate available", func(t *testing.T) {
		// Arrange - Create a transaction first
		createReq := map[string]interface{}{
			"description": "Test Purchase",
			"date":        "2024-01-15T10:30:00Z",
			"amount":      100.00,
		}
		jsonBody, _ := json.Marshal(createReq)

		createHttpReq := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createHttpReq)

		require.Equal(t, http.StatusCreated, createW.Code)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)
		transactionID := createResponse["id"].(string)

		// Act - Try to convert (should fail - no exchange rate)
		convertReq := map[string]interface{}{
			"target_currency": "BRL",
		}
		convertJsonBody, _ := json.Marshal(convertReq)

		convertHttpReq := httptest.NewRequest("POST", "/api/v1/transactions/"+transactionID+"/convert", bytes.NewBuffer(convertJsonBody))
		convertHttpReq.Header.Set("Content-Type", "application/json")
		convertW := httptest.NewRecorder()
		router.ServeHTTP(convertW, convertHttpReq)

		// Assert
		assert.Equal(t, http.StatusUnprocessableEntity, convertW.Code)

		var response map[string]interface{}
		err = json.Unmarshal(convertW.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Failed to convert transaction")
		assert.Contains(t, response["details"], "no suitable exchange rate found")
	})

	t.Run("Convert transaction - invalid currency", func(t *testing.T) {
		// Arrange - Create a transaction first
		createReq := map[string]interface{}{
			"description": "Test Purchase",
			"date":        "2024-01-15T10:30:00Z",
			"amount":      100.00,
		}
		jsonBody, _ := json.Marshal(createReq)

		createHttpReq := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(jsonBody))
		createHttpReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()
		router.ServeHTTP(createW, createHttpReq)

		require.Equal(t, http.StatusCreated, createW.Code)

		var createResponse map[string]interface{}
		err := json.Unmarshal(createW.Body.Bytes(), &createResponse)
		require.NoError(t, err)
		transactionID := createResponse["id"].(string)

		// Act - Try to convert to USD (should fail - USD to USD)
		convertReq := map[string]interface{}{
			"target_currency": "USD",
		}
		convertJsonBody, _ := json.Marshal(convertReq)

		convertHttpReq := httptest.NewRequest("POST", "/api/v1/transactions/"+transactionID+"/convert", bytes.NewBuffer(convertJsonBody))
		convertHttpReq.Header.Set("Content-Type", "application/json")
		convertW := httptest.NewRecorder()
		router.ServeHTTP(convertW, convertHttpReq)

		// Assert
		assert.Equal(t, http.StatusBadRequest, convertW.Code)

		var response map[string]interface{}
		err = json.Unmarshal(convertW.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Failed to convert transaction")
	})

	t.Run("Convert transaction - invalid UUID", func(t *testing.T) {
		// Arrange
		convertReq := map[string]interface{}{
			"target_currency": "BRL",
		}
		convertJsonBody, _ := json.Marshal(convertReq)

		// Act
		convertHttpReq := httptest.NewRequest("POST", "/api/v1/transactions/invalid-uuid/convert", bytes.NewBuffer(convertJsonBody))
		convertHttpReq.Header.Set("Content-Type", "application/json")
		convertW := httptest.NewRecorder()
		router.ServeHTTP(convertW, convertHttpReq)

		// Assert
		assert.Equal(t, http.StatusBadRequest, convertW.Code)

		var response map[string]interface{}
		err := json.Unmarshal(convertW.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response["error"], "Invalid transaction ID format")
	})
}

func TestHealthCheckAPI(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("Health check endpoint", func(t *testing.T) {
		// Act
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "purchase-transaction-api", response["service"])
	})
}

func TestAPIDocumentationEndpoint(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	t.Run("API documentation endpoint", func(t *testing.T) {
		// Act
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Purchase Transaction API", response["service"])
		assert.Equal(t, "1.0.0", response["version"])
		assert.NotEmpty(t, response["endpoints"])
	})
}
