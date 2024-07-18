package main

import (
	"fmt"
	"log"

	"go-processor/internal/config"
	"go-processor/internal/rabbitmq"
)

func main() {
	fmt.Println("Transaction Processor Starting...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Loaded configuration: %+v\n", cfg)

	rmq, err := rabbitmq.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()

	err = rmq.ConsumeMessages(processMessage)
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}
}

func processMessage(body []byte) error {
	// TODO: Implement actual message processing
	fmt.Printf("Processing message: %s\n", string(body))
	return nil
}