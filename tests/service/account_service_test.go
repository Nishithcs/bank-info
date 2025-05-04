// tests/service/account_service_test.go
package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Nishithcs/bank-info/internal/domain"
	"github.com/Nishithcs/bank-info/internal/queue"
	"github.com/Nishithcs/bank-info/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(ctx context.Context, account domain.Account) (domain.Account, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(domain.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByID(ctx context.Context, id string) (domain.Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Account), args.Error(1)
}

func (m *MockAccountRepository) UpdateBalance(ctx context.Context, id string, amount float64, transactionType domain.TransactionType) (domain.Account, error) {
	args := m.Called(ctx, id, amount, transactionType)
	return args.Get(0).(domain.Account), args.Error(1)
}

type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	args := m.Called(ctx, transaction)
	return args.Get(0).(domain.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]domain.Transaction), args.Error(1)
}

type MockQueueService struct {
	mock.Mock
}

func (m *MockQueueService) PublishAccountCreation(ctx context.Context, payload queue.AccountCreationPayload) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockQueueService) PublishTransaction(ctx context.Context, payload queue.TransactionPayload) (string, error) {
	args := m.Called(ctx, payload)
	return args.String(0), args.Error(1)
}

func (m *MockQueueService) Consume(queueName string, handler func([]byte) error) error {
	args := m.Called(queueName, handler)
	return args.Error(0)
}

func (m *MockQueueService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestCreateAccount(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAccountRepo := new(MockAccountRepository)
	mockTransactionRepo := new(MockTransactionRepository)
	mockQueueService := new(MockQueueService)

	accountService := service.NewAccountService(mockAccountRepo, mockTransactionRepo, mockQueueService)

	req := domain.CreateAccountRequest{
		Name:          "Test Account",
		InitialAmount: 1000.0,
	}

	expectedPayload := queue.AccountCreationPayload{
		Name:          req.Name,
		InitialAmount: req.InitialAmount,
	}

	// Expectations
	mockQueueService.On("PublishAccountCreation", ctx, expectedPayload).Return(nil)

	// Test
	err := accountService.CreateAccount(ctx, req)

	// Assert
	assert.NoError(t, err)
	mockQueueService.AssertExpectations(t)
}

func TestGetAccount(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAccountRepo := new(MockAccountRepository)
	mockTransactionRepo := new(MockTransactionRepository)
	mockQueueService := new(MockQueueService)

	accountService := service.NewAccountService(mockAccountRepo, mockTransactionRepo, mockQueueService)

	accountID := uuid.New().String()
	account := domain.Account{
		ID:        accountID,
		Name:      "Test Account",
		Balance:   1000.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)

	// Test
	result, err := accountService.GetAccount(ctx, accountID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, account.ID, result.ID)
	assert.Equal(t, account.Name, result.Name)
	assert.Equal(t, account.Balance, result.Balance)
	mockAccountRepo.AssertExpectations(t)
}

func TestProcessTransaction(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAccountRepo := new(MockAccountRepository)
	mockTransactionRepo := new(MockTransactionRepository)
	mockQueueService := new(MockQueueService)

	accountService := service.NewAccountService(mockAccountRepo, mockTransactionRepo, mockQueueService)

	accountID := uuid.New().String()
	account := domain.Account{
		ID:        accountID,
		Name:      "Test Account",
		Balance:   1000.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	req := domain.TransactionRequest{
		AccountID: accountID,
		Amount:    500.0,
		Type:      domain.Credit,
	}

	referenceID := uuid.New().String()

	// Expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockQueueService.On("PublishTransaction", ctx, mock.AnythingOfType("queue.TransactionPayload")).
		Return(referenceID, nil)

	// Test
	result, err := accountService.ProcessTransaction(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, referenceID, result.ReferenceID)
	assert.Equal(t, "pending", result.Status)
	mockAccountRepo.AssertExpectations(t)
	mockQueueService.AssertExpectations(t)
}

func TestGetTransactionHistory(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAccountRepo := new(MockAccountRepository)
	mockTransactionRepo := new(MockTransactionRepository)
	mockQueueService := new(MockQueueService)

	accountService := service.NewAccountService(mockAccountRepo, mockTransactionRepo, mockQueueService)

	accountID := uuid.New().String()
	account := domain.Account{
		ID:        accountID,
		Name:      "Test Account",
		Balance:   1000.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	transactions := []domain.Transaction{
		{
			ID:              uuid.New().String(),
			AccountID:       accountID,
			ReferenceID:     uuid.New().String(),
			Amount:          500.0,
			Type:            domain.Credit,
			PreviousBalance: 500.0,
			NewBalance:      1000.0,
			Timestamp:       time.Now(),
		},
	}

	// Expectations
	mockAccountRepo.On("GetByID", ctx, accountID).Return(account, nil)
	mockTransactionRepo.On("GetByAccountID", ctx, accountID).Return(transactions, nil)

	// Test
	result, err := accountService.GetTransactionHistory(ctx, accountID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, len(transactions), len(result.Transactions))
	mockAccountRepo.AssertExpectations(t)
	mockTransactionRepo.AssertExpectations(t)
}

func TestProcessAccountCreation(t *testing.T) {
	// Setup
	ctx := context.Background()
	mockAccountRepo := new(MockAccountRepository)
	mockTransactionRepo := new(MockTransactionRepository)
	mockQueueService := new(MockQueueService)

	accountService := service.NewAccountService(mockAccountRepo, mockTransactionRepo, mockQueueService)

	accountID := uuid.New().String()
	
	accountPayload := queue.AccountCreationPayload{
		Name:          "Test Account",
		InitialAmount: 1000.0,
	}
	
	task := queue.Task{
		ID:        uuid.New().String(),
		Type:      queue.CreateAccountTask,
		Payload:   mustMarshal(accountPayload),
		CreatedAt: time.Now(),
	}
	
	taskBytes := mustMarshal(task)
	
	expectedAccount := domain.Account{
		ID:        mock.AnythingOfType("string"),
		Name:      accountPayload.Name,
		Balance:   accountPayload.InitialAmount,
		CreatedAt: mock.AnythingOfType("time.Time"),
		UpdatedAt: mock.AnythingOfType("time.Time"),
	}
	
	expectedTransaction := domain.Transaction{
		ID:              mock.AnythingOfType("string"),
		AccountID:       mock.AnythingOfType("string"),
		ReferenceID:     mock.AnythingOfType("string"),
		Amount:          accountPayload.InitialAmount,
		Type:            domain.Credit,
		PreviousBalance: 0,
		NewBalance:      accountPayload.InitialAmount,
		Timestamp:       mock.