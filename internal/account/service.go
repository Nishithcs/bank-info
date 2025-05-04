package service

import (
	"context"
	"github.com/Nishithcs/bank-info/internal/account"
)

type AccountService struct {
	accountRepo *account.AccountRepository
}

func NewAccountService(ar *account.AccountRepository) *AccountService {
	return &AccountService{accountRepo: ar}
}

func (s *AccountService) CreateAccount(ctx context.Context, id string, name string, initialBalance float64) error {
	return s.accountRepo.CreateAccount(ctx, id, name, initialBalance)
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*account.Account, error) {
	return s.accountRepo.GetAccount(ctx, id)
}

func (s *AccountService) UpdateAccountBalance(ctx context.Context, id string, amount float64, transactionType string) error {
	return s.accountRepo.UpdateAccountBalance(ctx, id, amount, transactionType)
}