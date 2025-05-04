package db

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/Nishithcs/bank-info/config"
)

var Mongo *mongo.Database

func InitMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GetEnv("MONGO_URI", "")))
	if err != nil {
		return err
	}
	Mongo = client.Database("bankinfo")
	return nil
}