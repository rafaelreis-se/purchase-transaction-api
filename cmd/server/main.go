package main

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/application/usecases"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/config"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/database"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/infrastructure/http/handlers"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

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
	router := http.NewRouter(transactionHandler)
	ginRouter := router.SetupRoutes()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port[1:] // Remove ':' from config
	}

	// Initialize and start server
	server := http.NewServer(ginRouter, port)

	log.Printf("Purchase Transaction API starting on port %s", port)
	log.Println("Endpoints available:")
	log.Println("  GET  /health")
	log.Println("  GET  /")
	log.Println("  POST /api/v1/transactions")
	log.Println("  GET  /api/v1/transactions")
	log.Println("  GET  /api/v1/transactions/:id")
	log.Println("  POST /api/v1/transactions/:id/convert")

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
