package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initPaymentDB() error {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./payments.db"
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createPaymentTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	log.Printf("SQLite database initialized: %s", dbPath)
	return nil
}

func createPaymentTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS payments (
		id TEXT PRIMARY KEY,
		chain_id INTEGER NOT NULL,
		tx_hash TEXT,
		sender TEXT NOT NULL,
		sender_ens TEXT,
		recipient TEXT NOT NULL,
		recipient_ens TEXT,
		token TEXT NOT NULL,
		amount TEXT NOT NULL,
		is_private BOOLEAN DEFAULT FALSE,
		attestation_id TEXT,
		receipt_cid TEXT,
		metadata TEXT,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
	);

	CREATE INDEX IF NOT EXISTS idx_payments_sender ON payments(sender);
	CREATE INDEX IF NOT EXISTS idx_payments_recipient ON payments(recipient);
	CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
	CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments(created_at);
	CREATE INDEX IF NOT EXISTS idx_payments_chain_id ON payments(chain_id);

	CREATE TABLE IF NOT EXISTS receipts (
		id TEXT PRIMARY KEY,
		payment_id TEXT NOT NULL,
		receipt_data TEXT NOT NULL,
		storage_cid TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(payment_id) REFERENCES payments(id)
	);

	CREATE INDEX IF NOT EXISTS idx_receipts_payment_id ON receipts(payment_id);
	CREATE INDEX IF NOT EXISTS idx_receipts_created_at ON receipts(created_at);
	`

	_, err := db.Exec(schema)
	return err
}

func closeDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}