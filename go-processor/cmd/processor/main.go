package main

import (
	"fmt"
	"log"

	"go-processor/internal/config"
	// Uncomment these when implemented
	// "go-processor/internal/database"
	// "go-processor/internal/rabbitmq"
)

func main() {
	fmt.Println("Transaction Processor Starting...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Loaded configuration: %+v\n", cfg)

	// TODO: Initialize RabbitMQ connection
	// TODO: Initialize database connection
	// TODO: Start processing transactions
}