package transactions

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Transaction struct {
	Sender    string
	Recipient string
	Amount    float64
	Timestamp int64
	Signature string
}

func NewTransaction(sender, recipient string, amount float64) *Transaction {
	return &Transaction{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		Timestamp: time.Now().Unix(),
	}
}

func (tx *Transaction) CalculateHash() string {
	data := fmt.Sprintf("%s%s%f%d", tx.Sender, tx.Recipient, tx.Amount, tx.Timestamp)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

