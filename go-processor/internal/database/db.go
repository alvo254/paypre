package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

type Transaction struct {
	ID        int
	Sender    string
	Recipient string
	Amount    float64
	CreatedAt time.Time
}

type FailedTransaction struct {
	ID        int
	Sender    string
	Recipient string
	Amount    float64
	Error     string
	Attempts  int
	CreatedAt time.Time
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	if err = createTablesIfNotExist(db); err != nil {
		return nil, fmt.Errorf("error creating tables: %w", err)
	}

	return &DB{db}, nil
}

func createTablesIfNotExist(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			sender VARCHAR(255) NOT NULL,
			recipient VARCHAR(255) NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS failed_transactions (
			id SERIAL PRIMARY KEY,
			sender VARCHAR(255) NOT NULL,
			recipient VARCHAR(255) NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			error TEXT NOT NULL,
			attempts INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func (db *DB) InsertTransaction(sender, recipient string, amount float64) error {
	_, err := db.Exec("INSERT INTO transactions (sender, recipient, amount) VALUES ($1, $2, $3)",
		sender, recipient, amount)
	if err != nil {
		return fmt.Errorf("error inserting transaction: %w", err)
	}
	return nil
}

func (db *DB) InsertFailedTransaction(sender, recipient string, amount float64, errMsg string) error {
	_, err := db.Exec(`
		INSERT INTO failed_transactions (sender, recipient, amount, error, attempts)
		VALUES ($1, $2, $3, $4, 1)
	`, sender, recipient, amount, errMsg)
	if err != nil {
		return fmt.Errorf("error inserting failed transaction: %w", err)
	}
	return nil
}

func (db *DB) GetFailedTransactions() ([]FailedTransaction, error) {
	rows, err := db.Query(`
		SELECT id, sender, recipient, amount, error, attempts, created_at
		FROM failed_transactions
		WHERE attempts < 3
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying failed transactions: %w", err)
	}
	defer rows.Close()

	var transactions []FailedTransaction
	for rows.Next() {
		var t FailedTransaction
		err := rows.Scan(&t.ID, &t.Sender, &t.Recipient, &t.Amount, &t.Error, &t.Attempts, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning failed transaction: %w", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (db *DB) UpdateFailedTransaction(id int, success bool) error {
	if success {
		_, err := db.Exec("DELETE FROM failed_transactions WHERE id = $1", id)
		if err != nil {
			return fmt.Errorf("error deleting successful transaction: %w", err)
		}
	} else {
		_, err := db.Exec("UPDATE failed_transactions SET attempts = attempts + 1 WHERE id = $1", id)
		if err != nil {
			return fmt.Errorf("error updating failed transaction attempts: %w", err)
		}
	}
	return nil
}