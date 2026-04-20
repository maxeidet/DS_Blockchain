package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Transaction struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "genesis", "coinbase", "transfer"
}

func NewTransferTransaction(from, to string, amount int) Transaction {
	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
		Type:      "transfer",
	}
	tx.ID = CalculateTransactionID(tx)
	return tx
}

func NewGenesisTransaction(to string, amount int) Transaction {
	tx := Transaction{
		From:      "",
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
		Type:      "genesis",
	}
	tx.ID = CalculateTransactionID(tx)
	return tx
}

func NewGenesisTransactionWithTimestamp(to string, amount int, timestamp int64) Transaction {
	tx := Transaction{
		From:      "",
		To:        to,
		Amount:    amount,
		Timestamp: timestamp,
		Type:      "genesis",
	}
	tx.ID = CalculateTransactionID(tx)
	return tx
}

func NewCoinbaseTransaction(to string, amount int) Transaction {
	tx := Transaction{
		From:      "",
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
		Type:      "coinbase",
	}
	tx.ID = CalculateTransactionID(tx)
	return tx
}

func CalculateTransactionID(tx Transaction) string {
	record := fmt.Sprintf("%s|%s|%d|%d|%s", tx.From, tx.To, tx.Amount, tx.Timestamp, tx.Type)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}