package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RabbitMQURL            string
	PostgresURL            string
	MPesaConsumerKey       string
	MPesaConsumerSecret    string
	MPesaPassKey           string
	MPesaBusinessShortCode string
}

func Load() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	config := &Config{
		RabbitMQURL:            os.Getenv("RABBITMQ_URL"),
		PostgresURL:            os.Getenv("POSTGRES_URL"),
		MPesaConsumerKey:       os.Getenv("MPESA_CONSUMER_KEY"),
		MPesaConsumerSecret:    os.Getenv("MPESA_CONSUMER_SECRET"),
		MPesaPassKey:           os.Getenv("MPESA_PASS_KEY"),
		MPesaBusinessShortCode: os.Getenv("MPESA_BUSINESS_SHORTCODE"),
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL is not set")
	}
	if c.PostgresURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set")
	}
	if c.MPesaConsumerKey == "" {
		return fmt.Errorf("MPESA_CONSUMER_KEY is not set")
	}
	if c.MPesaConsumerSecret == "" {
		return fmt.Errorf("MPESA_CONSUMER_SECRET is not set")
	}
	if c.MPesaPassKey == "" {
		return fmt.Errorf("MPESA_PASS_KEY is not set")
	}
	if c.MPesaBusinessShortCode == "" {
		return fmt.Errorf("MPESA_BUSINESS_SHORTCODE is not set")
	}
	return nil
}