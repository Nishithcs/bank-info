//  internal/account/handler.go
package account

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/Nishithcs/bank-info/internal/mq"
	"github.com/Nishithcs/bank-info/internal/account/service"
	"github.com/google/uuid"
)

type AccountHandler struct {
	accountService *service.AccountService
	mqPublisher    *mq.Publisher
}

func NewAccountHandler(as *service.AccountService, mp *mq.Publisher) *AccountHandler {
	return &AccountHandler{accountService: as, mqPublisher: mp}
}

type CreateAccountRequest struct {
	AccountName  string  `json:"account_name" binding:"required"`
	InitialAmount float64 `json:"initial_amount" binding:"required"`
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accountID := uuid.New().String()

	//  Publish message to RabbitMQ
	err := h.mqPublisher.Publish("account.create", map[string]interface{}{
		"account_id":   accountID,
		"account_name": req.AccountName,
		"initial_amount": req.InitialAmount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish account creation message"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Account creation request submitted", "account_id": accountID})
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	accountID := c.Param("id")

	account, err := h.accountService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account"})
		return
	}

	c.JSON(http.StatusOK, account)
}