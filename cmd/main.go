package main

import (
	"github.com/gin-gonic/gin"
	"github.com/Nishithcs/bank-info/internal/api"
	"github.com/Nishithcs/bank-info/internal/queue"
	"github.com/Nishithcs/bank-info/config"
)

func main() {
	config.LoadEnv()
	queue.StartConsumer()
	r := gin.Default()
	api.SetupRoutes(r)
	r.Run(":8080")
}