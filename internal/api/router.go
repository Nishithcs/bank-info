package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/account", CreateAccount)
	r.POST("/transaction", HandleTransaction)
	r.GET("/account/:id", GetAccountInfo)
	r.GET("/transactions/:id", GetTransactionHistory)
}