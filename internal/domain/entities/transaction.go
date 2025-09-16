package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Transaction represents a purchase transaction in the system
type Transaction struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Description string    `json:"description" gorm:"not null" validate:"required,max=50"`
	Date        time.Time `json:"date" gorm:"not null" validate:"required"`
	Amount      Money     `json:"amount" gorm:"not null" validate:"required,gt=0"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// Money represents a monetary value in cents to avoid floating point precision issues
type Money int64

// NewMoney creates a Money value from dollars (converts to cents)
func NewMoney(dollars float64) Money {
	// Round to nearest cent and convert to int64 cents
	return Money(dollars*100 + 0.5)
}

// Dollars returns the monetary value in dollars (float64)
func (m Money) Dollars() float64 {
	return float64(m) / 100.0
}

// Cents returns the money value in cents
func (m Money) Cents() int64 {
	return int64(m)
}

// IsPositive checks if the money value is positive
func (m Money) IsPositive() bool {
	return m > 0
}

// Validate performs business rule validation
func (t *Transaction) Validate() error {
	if t.Description == "" {
		return fmt.Errorf("description is required")
	}

	if len(t.Description) > 50 {
		return fmt.Errorf("description must not exceed 50 characters")
	}

	if t.Date.IsZero() {
		return fmt.Errorf("transaction date is required")
	}

	if !t.Amount.IsPositive() {
		return fmt.Errorf("purchase amount must be positive")
	}

	return nil
}
