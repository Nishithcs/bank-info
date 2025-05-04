package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/Nishithcs/bank-info/internal/db"
	"github.com/Nishithcs/bank-info/internal/models"
)

type TransactionTask struct {
	AccountID int64
	Amount    float64
	Type      string // credit or debit
}

func ProcessTransaction(t TransactionTask) {
	tx := db.DB.MustBegin()
	if t.Type == "credit" {
		tx.MustExec("UPDATE accounts SET amount = amount + $1 WHERE id = $2", t.Amount, t.AccountID)
	} else if t.Type == "debit" {
		tx.MustExec("UPDATE accounts SET amount = amount - $1 WHERE id = $2", t.Amount, t.AccountID)
	}
	tx.Commit()
	trans := models.Transaction{
		AccountID:     t.AccountID,
		Amount:        t.Amount,
		Type:          t.Type,
		TransactionID: uuid.New().String(),
	}
	db.Mongo.Collection("transactions").InsertOne(context.TODO(), trans)
}
