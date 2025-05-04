package models

type Account struct {
	ID     int64   `db:"id" json:"id"`
	Name   string  `db:"name" json:"name"`
	Amount float64 `db:"amount" json:"amount"`
}

type Transaction struct {
	AccountID     int64   `bson:"account_id" json:"account_id"`
	Amount        float64 `bson:"amount" json:"amount"`
	TransactionID string  `bson:"transaction_id" json:"transaction_id"`
	Type          string  `bson:"type" json:"type"`
}