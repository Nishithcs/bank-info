package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Nishithcs/bank-info/internal/account"
	"github.com/Nishithcs/bank-info/internal/transaction"
	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"log"
	"time"
)

// AccountWorker handles messages related to account creation.
type AccountWorker struct {
	accountRepo *account.AccountRepository
}

// NewAccountWorker creates a new AccountWorker.
func NewAccountWorker(ar *account.AccountRepository) *AccountWorker {
	return &AccountWorker{accountRepo: ar}
}

// ProcessMessages consumes and processes account-related messages from RabbitMQ.
func (w *AccountWorker) ProcessMessages(messages <-chan amqp.Delivery) {
	ctx := context.Background() // Use a background context
	for msg := range messages {
		fmt.Printf("Received account creation message: %s\n", msg.Body)

		var accountData map[string]interface{}
		if err := json.Unmarshal(msg.Body, &accountData); err != nil {
			log.Printf("Error unmarshaling message: %v\n", err)
			msg.Nack(false, false) // Reject the message (don't requeue)
			continue
		}

		accountID, ok := accountData["account_id"].(string)
		if !ok {
			log.Println("account_id not found or not a string")
			msg.Nack(false, false)
			continue
		}
		accountName, ok := accountData["account_name"].(string)
		if !ok {
			log.Println("account_name not found or not a string")
			msg.Nack(false, false)
			continue
		}
		initialAmount, ok := accountData["initial_amount"].(float64)
		if !ok {
			log.Println("initial_amount not found or not a float64")
			msg.Nack(false, false)
			continue
		}

		err := w.accountRepo.CreateAccount(ctx, accountID, accountName, initialAmount)
		if err != nil {
			log.Printf("Error creating account: %v\n", err)
			msg.Nack(false, false) // Reject the message
			continue
		}

		fmt.Println("Account created successfully")
		msg.Ack(false) // Acknowledge the message
	}
}

// TransactionWorker handles messages related to deposit and withdrawal transactions.
type TransactionWorker struct {
	accountRepo     *account.AccountRepository
	transactionRepo *transaction.TransactionRepository
}

// NewTransactionWorker creates a new TransactionWorker.
func NewTransactionWorker(ar *account.AccountRepository, tr *transaction.TransactionRepository) *TransactionWorker {
	return &TransactionWorker{accountRepo: ar, transactionRepo: tr}
}

// ProcessMessages consumes and processes transaction-related messages from RabbitMQ.
func (w *TransactionWorker) ProcessMessages(messages <-chan amqp.Delivery) {
	ctx := context.Background()
	for msg := range messages {
		fmt.Printf("Received transaction message: %s\n", msg.Body)

		var transactionData map[string]interface{}
		if err := json.Unmarshal(msg.Body, &transactionData); err != nil {
			log.Printf("Error unmarshaling message: %v\n", err)
			msg.Nack(false, false)
			continue
		}

		accountID, ok := transactionData["account_id"].(string)
		if !ok {
			log.Println("account_id not found or not a string")
			msg.Nack(false, false)
			continue
		}
		amount, ok := transactionData["amount"].(float64)
		if !ok {
			log.Println("amount not found or not a float64")
			msg.Nack(false, false)
			continue
		}
		transactionType, ok := transactionData["transaction_type"].(string)
		if !ok || (transactionType != "credit" && transactionType != "debit") {
			log.Println("invalid transaction_type")
			msg.Nack(false, false)
			continue
		}

		referenceID := uuid.New().String()

		// Update account balance in PostgreSQL
		err := w.accountRepo.UpdateAccountBalance(ctx, accountID, amount, transactionType)
		if err != nil {
			log.Printf("Error updating account balance: %v\n", err)
			msg.Nack(false, false)
			continue
		}

		// Create transaction history entry
		transactionHistory := transaction.TransactionHistory{
			AccountID:     accountID,
			Amount:        amount,
			TransactionType: transactionType,
			ReferenceID:   referenceID,
			Timestamp:     time.Now(),
		}
		err = w.transactionRepo.CreateTransactionHistory(ctx, &transactionHistory)
		if err != nil {
			log.Printf("Error creating transaction history: %v\n", err)
			msg.Nack(false, false)
			continue
		}

		fmt.Println("Transaction processed successfully")
		msg.Ack(false)
	}
}