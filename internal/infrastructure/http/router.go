package http

import (
	"github.com/gin-gonic/gin"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http/handlers"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http/middleware"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/pkg/logger"
)

// Router sets up the HTTP routes for the application
type Router struct {
	transactionHandler *handlers.TransactionHandler
	logger             *logger.Logger
}

// NewRouter creates a new Router with the provided handlers
func NewRouter(transactionHandler *handlers.TransactionHandler, log *logger.Logger) *Router {
	return &Router{
		transactionHandler: transactionHandler,
		logger:             log,
	}
}

// SetupRoutes configures all the routes for the application
func (r *Router) SetupRoutes() *gin.Engine {
	// Create Gin router without default logger (we'll use our structured logger)
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add custom middleware with structured logging
	router.Use(middleware.RequestIDMiddleware(r.logger))
	router.Use(middleware.LoggingMiddleware(r.logger))
	router.Use(middleware.ErrorLoggingMiddleware(r.logger))
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	// Health check endpoint for Docker
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "healthy",
			"service":   "purchase-transaction-api",
			"timestamp": gin.H{"unix": gin.H{}},
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Transaction routes
		transactions := v1.Group("/transactions")
		{
			// POST /api/v1/transactions - Create a new transaction
			transactions.POST("", r.transactionHandler.CreateTransaction)

			// GET /api/v1/transactions - List transactions with pagination
			transactions.GET("", r.transactionHandler.ListTransactions)

			// GET /api/v1/transactions/:id - Get a specific transaction
			transactions.GET("/:id", r.transactionHandler.GetTransaction)

			// POST /api/v1/transactions/:id/convert - Convert transaction currency
			transactions.POST("/:id/convert", r.transactionHandler.ConvertTransaction)
		}
	}

	// API documentation endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Purchase Transaction API",
			"version": "1.0.0",
			"endpoints": gin.H{
				"health": "GET /health",
				"transactions": gin.H{
					"create":  "POST /api/v1/transactions",
					"list":    "GET /api/v1/transactions?page=1&size=20",
					"get":     "GET /api/v1/transactions/{id}",
					"convert": "POST /api/v1/transactions/{id}/convert",
				},
			},
		})
	})

	return router
}
