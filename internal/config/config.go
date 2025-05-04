// internal/config/config.go
package config

// Config contains application configuration
type Config struct {
	ServerPort  int    `json:"server_port"`
	PostgresURL string `json:"postgres_url"`
	MongoURI    string `json:"mongo_uri"`
	MongoDB     string `json:"mongo_db"`
	RabbitMQURL string `json:"rabbitmq_url"`
}