package config

import (
	"os"
)

type Config struct {
	RabbitMQURL string
	PostgresURL string
}

func Load() (*Config, error) {
	return &Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		PostgresURL: getEnv("POSTGRES_URL", "postgres://admin:adminpassword@localhost:5432/saasfinance?sslmode=disable"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}