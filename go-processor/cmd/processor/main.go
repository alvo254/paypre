package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-processor/internal/config"
	"go-processor/internal/database"
	"go-processor/internal/rabbitmq"
)

type TransactionMessage struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
}

func main() {
	fmt.Println("Transaction Processor Starting...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Loaded configuration: %+v\n", cfg)

	db, err := database.NewDB(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	rmq, err := rabbitmq.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()

	fmt.Println("Successfully connected to RabbitMQ and database")

	// Start the retry mechanism
	go retryFailedTransactions(db)

	err = rmq.ConsumeMessages(func(body []byte) error {
		return processMessage(db, body)
	})
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}
}

func processMessage(db *database.DB, body []byte) error {
	var msg TransactionMessage
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	log.Printf("Received transaction: %+v", msg)

	err = db.InsertTransaction(msg.Sender, msg.Recipient, msg.Amount)
	if err != nil {
		log.Printf("Failed to process transaction: %v", err)
		errInsert := db.InsertFailedTransaction(msg.Sender, msg.Recipient, msg.Amount, err.Error())
		if errInsert != nil {
			log.Printf("Failed to insert failed transaction: %v", errInsert)
		}
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	log.Println("Transaction processed successfully")
	return nil
}

func retryFailedTransactions(db *database.DB) {
	for {
		failedTransactions, err := db.GetFailedTransactions()
		if err != nil {
			log.Printf("Error getting failed transactions: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, ft := range failedTransactions {
			err := db.InsertTransaction(ft.Sender, ft.Recipient, ft.Amount)
			if err != nil {
				log.Printf("Retry failed for transaction %d: %v", ft.ID, err)
				db.UpdateFailedTransaction(ft.ID, false)
			} else {
				log.Printf("Successfully retried transaction %d", ft.ID)
				db.UpdateFailedTransaction(ft.ID, true)
			}
		}

		time.Sleep(5 * time.Minute)
	}
}