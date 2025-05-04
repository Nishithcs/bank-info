package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	API      APIConfig      `mapstructure:"api"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Mongo    MongoConfig    `mapstructure:"mongo"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

type APIConfig struct {
	Port string `mapstructure:"port"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type MongoConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RabbitMQConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	AccountQueue    string `mapstructure:"account_queue"`
	TransactionQueue string `mapstructure:"transaction_queue"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.SetConfigName("config") //  Name of config file (without extension)

	viper.AutomaticEnv() //  Read in environment variables that match
	viper.SetEnvPrefix("BANK") //  Prefix for env vars to avoid collision

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(*viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found; using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}