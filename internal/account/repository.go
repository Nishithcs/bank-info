//  internal/account/repository.go
package account

import (
	"context"
	"github.com/Nishithcs/bank-info/internal/database"
	"gorm.io/gorm"
	"errors"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

type Account struct {
	ID      string  `gorm:"primaryKey"`
	Name    string
	Balance float64
}

func (r *AccountRepository) CreateAccount(ctx context.Context, id string, name string, initialBalance float64) error {
	account := Account{ID: id, Name: name, Balance: initialBalance}
	result := r.db.WithContext(ctx).Create(&account)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *AccountRepository) GetAccount(ctx context.Context, id string) (*Account, error) {
	var account Account
	result := r.db.WithContext(ctx).First(&account, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, result.Error
	}
	return &account, nil
}

func (r *AccountRepository) UpdateAccountBalance(ctx context.Context, id string, amount float64, transactionType string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	account := Account{ID: id}
	if err := tx.WithContext(ctx).Lock(&account).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.WithContext(ctx).First(&account).Error; err != nil {
		tx.Rollback()
		return err
	}

	if transactionType == "debit" && account.Balance < amount {
		tx.Rollback()
		return errors.New("insufficient funds")
	}

	if transactionType == "credit" {
		account.Balance += amount
	} else if transactionType == "debit" {
		account.Balance -= amount
	}

	if err := tx.WithContext(ctx).Save(&account).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}