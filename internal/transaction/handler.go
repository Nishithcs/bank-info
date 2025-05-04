package transaction

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Nishithcs/bank-info/internal/mq"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	mqPublisher *mq.Publisher
}

func NewTransactionHandler(mp *mq.Publisher) *TransactionHandler {
	return &TransactionHandler{mqPublisher: mp}
}

type TransactionRequest struct {
	AccountID     string  `json:"account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	TransactionType string  `json:"transaction_type" binding:"required,oneof=credit debit"`
}

func (h *TransactionHandler) HandleTransaction(c *gin.Context) {
	var req TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	referenceID := uuid.New().String()

	// Publish message to RabbitMQ
	err := h.mqPublisher.Publish("transaction.process", map[string]interface{}{
		"account_id":      req.AccountID,
		"amount":          req.Amount,
		"transaction_type": req.TransactionType,
		"reference_id":    referenceID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish transaction message"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Transaction request submitted", "reference_id": referenceID})
}

func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	accountID := c.Param("accountID")

	//  TODO: Implement GetTransactionHistory in service and repository
	//  For now, return a placeholder
	c.JSON(http.StatusOK, gin.H{"message": "Transaction history endpoint", "account_id": accountID})
}