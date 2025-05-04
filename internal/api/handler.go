package api

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/Nishithcs/bank-info/internal/db"
	"github.com/Nishithcs/bank-info/internal/queue"
	"github.com/Nishithcs/bank-info/internal/models"
	"github.com/Nishithcs/bank-info/internal/service"
	"context"
)

type AccountInput struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type TransactionInput struct {
	AccountID int64   `json:"account_id"`
	Amount    float64 `json:"amount"`
	Type      string  `json:"type"`
}

func CreateAccount(c *gin.Context) {
	var input AccountInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.DB.Exec("INSERT INTO accounts (name, amount) VALUES ($1, $2)", input.Name, input.Amount)
	c.JSON(http.StatusOK, gin.H{"status": "account creation task queued"})
}

func HandleTransaction(c *gin.Context) {
	var input TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refID := strconv.Itoa(int(input.AccountID)) + "-" + input.Type
	queue.Publish(service.TransactionTask{
		AccountID: input.AccountID,
		Amount:    input.Amount,
		Type:      input.Type,
	})
	c.JSON(http.StatusOK, gin.H{"reference_id": refID})
}

func GetAccountInfo(c *gin.Context) {
	id := c.Param("id")
	var acc models.Account
	db.DB.Get(&acc, "SELECT * FROM accounts WHERE id = $1", id)
	c.JSON(http.StatusOK, acc)
}

func GetTransactionHistory(c *gin.Context) {
	id := c.Param("id")
	var txns []models.Transaction
	cur, _ := db.Mongo.Collection("transactions").Find(context.TODO(), map[string]interface{}{"account_id": id})
	cur.All(context.TODO(), &txns)
	c.JSON(http.StatusOK, txns)
}
