package mocks

import (
	"time"

	"github.com/google/uuid"
	"github.com/rafaelreis-se/purchase-transaction-api/internal/domain/entities"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Save(transaction *entities.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(id uuid.UUID) (*entities.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetAll() ([]entities.Transaction, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetAllPaginated(page, size int) ([]entities.Transaction, int64, error) {
	args := m.Called(page, size)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entities.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) Update(transaction *entities.Transaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTransactionRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func (m *MockTransactionRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

// MockExchangeRateRepository is a mock implementation of ExchangeRateRepository
type MockExchangeRateRepository struct {
	mock.Mock
}

func (m *MockExchangeRateRepository) Save(exchangeRate *entities.ExchangeRate) error {
	args := m.Called(exchangeRate)
	return args.Error(0)
}

func (m *MockExchangeRateRepository) GetByID(id uuid.UUID) (*entities.ExchangeRate, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExchangeRate), args.Error(1)
}

func (m *MockExchangeRateRepository) FindRateForConversion(from, to entities.CurrencyCode, transactionDate time.Time) (*entities.ExchangeRate, error) {
	args := m.Called(from, to, transactionDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExchangeRate), args.Error(1)
}

func (m *MockExchangeRateRepository) Update(exchangeRate *entities.ExchangeRate) error {
	args := m.Called(exchangeRate)
	return args.Error(0)
}

func (m *MockExchangeRateRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockExchangeRateRepository) Exists(id uuid.UUID) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

// MockTreasuryService is a mock implementation of TreasuryService
type MockTreasuryService struct {
	mock.Mock
}

func (m *MockTreasuryService) FetchExchangeRate(from, to entities.CurrencyCode, date time.Time) (*entities.ExchangeRate, error) {
	args := m.Called(from, to, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ExchangeRate), args.Error(1)
}
