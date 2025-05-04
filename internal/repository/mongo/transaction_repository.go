// internal/repository/mongo/transaction_repository.go
package mongo

import (
	"context"
	"time"

	"github.com/Nishithcs/bank-info/internal/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository interface {
	Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error)
	GetByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error)
}

type transactionRepository struct {
	collection *mongo.Collection
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *mongo.Database) TransactionRepository {
	collection := db.Collection("transactions")
	
	// Create indexes for efficient queries
	_, err := collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "account_id", Value: 1}},
			Options: options.Index().SetBackground(true),
		},
	)
	
	if err != nil {
		panic(err)
	}
	
	_, err = collection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "reference_id", Value: 1}},
			Options: options.Index().SetBackground(true).SetUnique(true),
		},
	)
	
	if err != nil {
		panic(err)
	}

	return &transactionRepository{collection: collection}
}

// Create adds a new transaction to the ledger
func (r *transactionRepository) Create(ctx context.Context, transaction domain.Transaction) (domain.Transaction, error) {
	if transaction.ID == "" {
		transaction.ID = uuid.New().String()
	}
	
	if transaction.Timestamp.IsZero() {
		transaction.Timestamp = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, transaction)
	if err != nil {
		return domain.Transaction{}, err
	}

	return transaction, nil
}

// GetByAccountID retrieves transactions for an account
func (r *transactionRepository) GetByAccountID(ctx context.Context, accountID string) ([]domain.Transaction, error) {
	filter := bson.M{"account_id": accountID}
	
	options := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []domain.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}