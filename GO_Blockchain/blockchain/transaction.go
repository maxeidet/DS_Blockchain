package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Transaction struct {
	ID        string `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"` // "genesis", "coinbase", "transfer"`
	Nonce     uint64 `json:"nonce"`

	PublicKey string `json:"public_key,omitempty"`
	Signature string `json:"signature,omitempty"`
}

func NewTransferTransaction(from, to string, amount int, nonce uint64) Transaction {
	tx := Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
		Type:      "transfer",
		Nonce:     nonce,
	}
	tx.ID = CalculateTransactionID(tx)
	return tx
}

func NewSignedTransferTransaction(wallet *Wallet, to string, amount int, nonce uint64) (Transaction, error) {
	tx := Transaction{
		From:      wallet.Address,
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
		Type:      "transfer",
		Nonce:     nonce,
		PublicKey: hex.EncodeToString(wallet.PublicKey),
	}

	signBytes, err := tx.SigningBytes()
	if err != nil {
		return Transaction{}, err
	}

	sig, err := SignBytes(wallet.PrivateKey, signBytes)
	if err != nil {
		return Transaction{}, err
	}

	tx.Signature = sig
	tx.ID = CalculateTransactionID(tx)
	return tx, nil
}

func NewGenesisTransaction(to string, amount int) Transaction {
	tx := Transaction{
		From:      "",
		To:        to,
		Amount:    amount,
		Timestamp: time.Now().UnixMilli(),
		Type:      "genesis",
		Nonce:     0,
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
		Nonce:     0,
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
		Nonce:     0,
	}
	tx.ID = CalculateTransactionID(tx)
	return tx
}

func (tx Transaction) SigningBytes() ([]byte, error) {
	payload := map[string]any{
		"from":      tx.From,
		"to":        tx.To,
		"amount":    tx.Amount,
		"timestamp": tx.Timestamp,
		"type":      tx.Type,
		"nonce":     tx.Nonce,
	}

	return json.Marshal(payload)
}

func CalculateTransactionID(tx Transaction) string {
	record := fmt.Sprintf(
		"%s|%s|%d|%d|%s|%d|%s|%s",
		tx.From,
		tx.To,
		tx.Amount,
		tx.Timestamp,
		tx.Type,
		tx.Nonce,
		tx.PublicKey,
		tx.Signature,
	)

	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}