// internal/repository/postgres/account_repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Nishithcs/bank-info/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// AccountRepository handles database operations for accounts
type AccountRepository interface {
	Create(ctx context.Context, account domain.Account) (domain.Account, error)
	GetByID(ctx context.Context, id string) (domain.Account, error)
	UpdateBalance(ctx context.Context, id string, amount float64, transactionType domain.TransactionType) (domain.Account, error)
}

type accountRepository struct {
	db *sqlx.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *sqlx.DB) AccountRepository {
	return &accountRepository{db: db}
}

// Create creates a new account
func (r *accountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	query := `
		INSERT INTO accounts (id, name, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, balance, created_at, updated_at
	`

	err := r.db.QueryRowxContext(
		ctx,
		query,
		account.ID,
		account.Name,
		account.Balance,
		account.CreatedAt,
		account.UpdatedAt,
	).StructScan(&account)

	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}

// GetByID retrieves an account by ID
func (r *accountRepository) GetByID(ctx context.Context, id string) (domain.Account, error) {
	var account domain.Account

	query := `
		SELECT id, name, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &account, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}
		return domain.Account{}, err
	}

	return account, nil
}

// UpdateBalance updates the account balance with proper locking to ensure consistency
func (r *accountRepository) UpdateBalance(ctx context.Context, id string, amount float64, transactionType domain.TransactionType) (domain.Account, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return domain.Account{}, err
	}
	defer tx.Rollback()

	// Lock the row for update to prevent concurrent modifications
	query := `
		SELECT id, name, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	var account domain.Account
	err = tx.GetContext(ctx, &account, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}
		return domain.Account{}, err
	}

	var newBalance float64
	if transactionType == domain.Credit {
		newBalance = account.Balance + amount
	} else {
		newBalance = account.Balance - amount
		if newBalance < 0 {
			return domain.Account{}, domain.ErrInsufficientBalance
		}
	}

	updateQuery := `
		UPDATE accounts
		SET balance = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, name, balance, created_at, updated_at
	`

	now := time.Now()
	err = tx.GetContext(
		ctx,
		&account,
		updateQuery,
		newBalance,
		now,
		id,
	)
	if err != nil {
		return domain.Account{}, err
	}

	if err = tx.Commit(); err != nil {
		return domain.Account{}, err
	}

	return account, nil
}