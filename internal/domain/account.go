// internal/domain/account.go
package domain

import (
	"errors"
	"time"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	Credit TransactionType = "credit"
	Debit  TransactionType = "debit"
)

// Account represents a bank account
type Account struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Balance   float64   `json:"balance" db:"balance"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Transaction represents a transaction in the ledger
type Transaction struct {
	ID            string         `json:"id" bson:"_id"`
	AccountID     string         `json:"account_id" bson:"account_id"`
	ReferenceID   string         `json:"reference_id" bson:"reference_id"`
	Amount        float64        `json:"amount" bson:"amount"`
	Type          TransactionType `json:"type" bson:"type"`
	PreviousBalance float64      `json:"previous_balance" bson:"previous_balance"`
	NewBalance    float64        `json:"new_balance" bson:"new_balance"`
	Timestamp     time.Time      `json:"timestamp" bson:"timestamp"`
}

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	Name          string  `json:"name" validate:"required"`
	InitialAmount float64 `json:"initial_amount" validate:"required,min=0"`
}

// TransactionRequest represents a deposit or withdrawal request
type TransactionRequest struct {
	AccountID string         `json:"account_id" validate:"required"`
	Amount    float64        `json:"amount" validate:"required,gt=0"`
	Type      TransactionType `json:"type" validate:"required,oneof=credit debit"`
}

// AccountResponse represents the response for account information
type AccountResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

// TransactionResponse represents the response after a transaction request
type TransactionResponse struct {
	ReferenceID string `json:"reference_id"`
	Status      string `json:"status"`
}

// TransactionHistoryResponse represents the response for transaction history
type TransactionHistoryResponse struct {
	Transactions []Transaction `json:"transactions"`
}

// Error types
var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrDuplicateAccount    = errors.New("account already exists")
	ErrTransactionFailed   = errors.New("transaction failed")
)