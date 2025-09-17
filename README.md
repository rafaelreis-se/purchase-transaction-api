# Purchase Transaction API

A Go application for managing purchase transactions with multi-currency conversion support using Treasury exchange rates.

## Project Overview

This application was built to meet the following requirements:

- Store purchase transactions with validation
- Convert transactions to different currencies using Treasury Reporting Rates of Exchange API
- Provide RESTful API endpoints for transaction management

## Technology Stack

After analyzing the project requirements, I chose the following packages:

### HTTP Framework

```bash
go get github.com/gin-gonic/gin
```

**Justification**: Gin is the most popular HTTP framework in Go, offering excellent performance, middleware support, and clean routing. Perfect for building production-ready REST APIs.

### Database & ORM

```bash
go get gorm.io/gorm
go get gorm.io/driver/sqlite
```

**Justification**:

- **GORM**: Industry standard ORM for Go, provides type-safe database operations and migration support
- **SQLite**: Meets the requirement of "fully functional without installing separate databases", while being production-ready and ACID compliant

### Testing

```bash
go get github.com/stretchr/testify
```

**Justification**: De facto standard for testing in Go. Provides assertions, test suites, and mocking capabilities essential for the required "functional automated testing for Production applications".

### Validation

```bash
go get github.com/go-playground/validator/v10
```

**Justification**: Robust validation library that integrates seamlessly with struct tags. Essential for validating transaction fields (description length, positive amounts, date formats).

### Configuration Management

```bash
go get github.com/spf13/viper
```

**Justification**: Popular configuration management library supporting multiple formats (JSON, YAML, ENV). Enables flexible configuration for different environments.

### UUID Generation

```bash
go get github.com/google/uuid
```

**Justification**: Google's UUID implementation provides cryptographically secure unique identifiers for transaction records, meeting the "unique identifier" requirement.

## 🏗 Architecture

This project follows Clean Architecture principles with clear separation of concerns:

```
├── cmd/server/          # Application entry point
├── internal/domain/     # Business logic and entities
├── internal/application/# Use cases and handlers
├── internal/infrastructure/ # Database, HTTP, external APIs
└── tests/              # Comprehensive test suite
```

## Getting Started

```bash
# Clone the repository
git clone https://github.com/rafaelreis-se/purchase-transaction-api.git
cd purchase-transaction-api

# Install dependencies
go mod tidy

# Run the application
go run cmd/server/main.go
```

## Development Status

- [x] Initial project setup
- [x] Domain entities and validation
- [x] Database layer implementation
- [ ] Use Cases
- [ ] REST API endpoints
- [ ] Treasury API integration
- [ ] Currency conversion logic
- [ ] Comprehensive test suite
- [ ] Documentation and examples

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```
