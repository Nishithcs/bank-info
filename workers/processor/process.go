package processor

import (
	"time"

	internal "github.com/Nishithcs/bank-info/pkg"
)

type ProcessWorker struct {
	PgxConn internal.PgDBConnection
	EsConn  internal.ElasticsearchClient
}

// Log the account creation transaction to Elasticsearch
type TransactionDocument struct {
	TransactionID           string    `json:"transaction_id"`
	AccountNumber           string    `json:"account_number"`
	Amount                  float64   `json:"amount"`
	Type                    string    `json:"type"`
	Status                  string    `json:"status"`
	Timestamp               time.Time `json:"timestamp"`
	BranchCode              string    `json:"branch_code"`
	BalanceAfterTransaction float64   `json:"balance_after_transaction"`
}