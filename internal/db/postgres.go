package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"github.com/Nishithcs/bank-info/config"
)

var DB *sqlx.DB

func InitPostgres() error {
	dsn := config.GetEnv("POSTGRES_DSN", "")
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return err
	}
	DB = db
	return nil
}