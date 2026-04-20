package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
	Nonce        int64         `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
}

func NewBlock(index int, transactions []Transaction, prevHash string, difficulty int) Block {
	return Block{
		Index:        index,
		Timestamp:    time.Now().UnixMilli(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Nonce:        0,
		Difficulty:   difficulty,
	}
}

func CalculateHash(block Block) string {
	txBytes, _ := json.Marshal(block.Transactions)

	record := fmt.Sprintf("%d|%d|%s|%s|%d|%d",
		block.Index,
		block.Timestamp,
		string(txBytes),
		block.PrevHash,
		block.Nonce,
		block.Difficulty,
	)

	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

func (b Block) Pretty() string {
	out, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Sprintf("Block{Index:%d Hash:%s}", b.Index, b.Hash)
	}
	return string(out)
}