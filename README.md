# Purchase Transaction API

A REST API application that stores purchase transactions in USD and converts them to other currencies using US Treasury exchange rates.

## What it does

- **Store purchase transactions** with description, date, and USD amount
- **Convert transactions** to foreign currencies (EUR, BRL, CAD, JPY, etc.)
- **Uses official exchange rates** from US Treasury API
- **Validates all data** according to business rules
- **Provides REST endpoints** for integration

Built to demonstrate clean architecture, testing practices, and production-ready deployment.

## Prerequisites

### Option 1: Local Development

- **Go 1.25+** - [Install Go](https://golang.org/doc/install)
- **Make** - Usually pre-installed on Linux/Mac
- **Git** - For cloning the repository

### Option 2: Docker Only

- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **No Go/Make needed** - Everything runs in container

## Quick Start

### Local Development

```bash
git clone https://github.com/rafaelreis-se/purchase-transaction-api.git
cd purchase-transaction-api

# Setup environment
cp .env.example .env

# Install dependencies and run
make run

# Test the API
make test
make api-test
```

### Docker (No Go Required)

```bash
git clone https://github.com/rafaelreis-se/purchase-transaction-api.git
cd purchase-transaction-api

# Setup environment
cp .env.example .env

# Build and run in Docker
make docker
```

make docker

````

**API ready at:** `http://localhost:8080`

## Available Commands

```bash
make help      # Show all available commands
make run       # Run application locally
make test      # Run all tests (236 tests)
make api-test  # Test complete API workflow
make docker    # Build and run in Docker
make health    # Check application status
````

## API Endpoints

### Store Transaction

```http
POST /api/v1/transactions
Content-Type: application/json

{
  "description": "Coffee purchase",
  "date": "2024-01-15T10:30:00Z",
  "amount": 25.50
}
```

### Convert Currency

```http
POST /api/v1/transactions/{id}/convert
Content-Type: application/json

{
  "target_currency": "EUR"
}
```

### Get Transaction

```http
GET /api/v1/transactions/{id}
GET /api/v1/transactions
```

## Supported Currencies

**Available:** EUR, BRL, CAD, JPY, CNY, AUD  
**Source:** US Treasury Reporting Rates API  
**Rule:** Uses exchange rate ≤ purchase date within 6 months

## Testing

**Import API Collection:** [`docs/insomnia-collection.json`](docs/insomnia-collection.json)

- Complete test scenarios included
- Error cases covered
- Ready to import into Insomnia

## Validation Rules

- **Description:** ≤ 50 characters
- **Date:** Valid ISO format
- **Amount:** Positive USD, rounded to cents
- **Currency conversion:** Must have rate within 6 months

## Architecture

- **Framework:** Gin + GORM
- **Database:** SQLite (embedded)
- **External API:** US Treasury exchange rates
- **Logging:** Structured JSON logging
- **Container:** Multi-stage Docker build (43MB)
