package main

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/config"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/database"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/external"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http/handlers"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/pkg/logger"
)

func main() {
	// Load .env file (ignore error if file doesn't exist - for production flexibility)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize structured logger
	appLogger := logger.NewLogger(logger.LoggerConfig{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})

	appLogger.Info("Starting Purchase Transaction API",
		"version", "1.0.0",
		"environment", os.Getenv("ENVIRONMENT"),
		"log_level", cfg.Logger.Level,
	)

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		appLogger.LogError(err, "Failed to initialize database")
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			appLogger.LogError(err, "Error closing database")
		}
	}()

	appLogger.Info("Database initialized successfully", "path", cfg.Database.Path)

	// Initialize repositories
	transactionRepo := database.NewTransactionRepository(db.GetDB())
	exchangeRateRepo := database.NewExchangeRateRepository(db.GetDB())

	// Initialize external services
	treasuryService := external.NewTreasuryAPIClient(&cfg.Treasury)
	appLogger.Info("External services initialized")

	// Initialize validator
	validator := validator.New()

	// Initialize use cases with logger context
	createTransactionUseCase := usecases.NewCreateTransactionUseCase(transactionRepo, validator)
	getTransactionUseCase := usecases.NewGetTransactionUseCase(transactionRepo)
	listTransactionsUseCase := usecases.NewListTransactionsUseCase(transactionRepo, validator)
	convertTransactionUseCase := usecases.NewConvertTransactionUseCase(transactionRepo, exchangeRateRepo, treasuryService, validator)

	appLogger.Info("Use cases initialized")

	// Initialize handlers
	transactionHandler := handlers.NewTransactionHandler(
		createTransactionUseCase,
		getTransactionUseCase,
		listTransactionsUseCase,
		convertTransactionUseCase,
	)

	// Initialize router with logger
	router := http.NewRouter(transactionHandler, appLogger)
	ginRouter := router.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port[1:] // Remove ':' from config
	}

	// Initialize and start server
	server := http.NewServer(ginRouter, port)

	appLogger.Info("Purchase Transaction API starting",
		"port", port,
		"endpoints", []string{
			"GET  /health",
			"GET  /",
			"POST /api/v1/transactions",
			"GET  /api/v1/transactions",
			"GET  /api/v1/transactions/:id",
			"POST /api/v1/transactions/:id/convert",
		},
	)

	if err := server.Start(); err != nil {
		appLogger.LogError(err, "Failed to start server")
		log.Fatalf("Failed to start server: %v", err)
	}
}
