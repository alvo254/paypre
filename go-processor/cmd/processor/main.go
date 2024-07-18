package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-processor/internal/config"
	"go-processor/internal/database"
	"go-processor/internal/mpesa"
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

	fmt.Println("Configuration loaded successfully")

	db, err := database.NewDB(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	mpesaService := mpesa.NewMPesa(mpesa.Config{
		ConsumerKey:       cfg.MPesaConsumerKey,
		ConsumerSecret:    cfg.MPesaConsumerSecret,
		PassKey:           cfg.MPesaPassKey,
		BusinessShortCode: cfg.MPesaBusinessShortCode,
		Environment:       "sandbox", // Change to "production" for live environment
	})
	log.Printf("M-Pesa service initialized with Business Short Code: %s", cfg.MPesaBusinessShortCode)

	rmq, err := rabbitmq.NewRabbitMQ(cfg.RabbitMQURL, "transactions")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()

	fmt.Println("Successfully connected to RabbitMQ and database")

	// Start the retry mechanism
	go retryFailedTransactions(db, mpesaService)

	err = rmq.ConsumeMessages(func(body []byte) error {
		return processMessage(db, mpesaService, body)
	})
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}
}

func processMessage(db *database.DB, mpesaService *mpesa.MPesa, body []byte) error {
	var msg TransactionMessage
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return fmt.Errorf("failed to parse message: %w", err)
	}

	log.Printf("Received transaction: Sender: %s, Recipient: %s, Amount: %.2f", msg.Sender, msg.Recipient, msg.Amount)

	// Initiate M-Pesa transaction
	response, err := mpesaService.InitiateSTKPush(msg.Sender, int(msg.Amount))
	if err != nil {
		log.Printf("Failed to initiate M-Pesa transaction: %v", err)
		log.Printf("M-Pesa error details: %+v", err)
		errInsert := db.InsertFailedTransaction(msg.Sender, msg.Recipient, msg.Amount, err.Error())
		if errInsert != nil {
			log.Printf("Failed to insert failed transaction: %v", errInsert)
		}
		return fmt.Errorf("failed to initiate M-Pesa transaction: %w", err)
	}

	// Enhanced logging of the response
	log.Printf("M-Pesa transaction initiated successfully. Response: %+v", response)

	// Store transaction with checkout request ID
	id, err := db.InsertTransaction(msg.Sender, msg.Recipient, msg.Amount, response.CheckoutRequestID)
	if err != nil {
		log.Printf("Failed to insert transaction: %v", err)
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	log.Printf("Transaction processed successfully. ID: %d, Checkout Request ID: %s", id, response.CheckoutRequestID)
	return nil
}

func retryFailedTransactions(db *database.DB, mpesaService *mpesa.MPesa) {
	for {
		failedTransactions, err := db.GetFailedTransactions()
		if err != nil {
			log.Printf("Error getting failed transactions: %v", err)
			time.Sleep(1 * time.Minute)
			continue
		}

		for _, ft := range failedTransactions {
			response, err := mpesaService.InitiateSTKPush(ft.Sender, int(ft.Amount))
			if err != nil {
				log.Printf("Retry failed for transaction %d: %v", ft.ID, err)
				db.UpdateFailedTransaction(ft.ID, false)
			} else {
				log.Printf("Successfully retried transaction %d", ft.ID)
				id, err := db.InsertTransaction(ft.Sender, ft.Recipient, ft.Amount, response.CheckoutRequestID)
				if err != nil {
					log.Printf("Failed to insert retried transaction: %v", err)
				} else {
					log.Printf("Retried transaction inserted successfully. ID: %d, Checkout Request ID: %s", id, response.CheckoutRequestID)
					db.UpdateFailedTransaction(ft.ID, true)
				}
			}
		}

		time.Sleep(5 * time.Minute)
	}
}
