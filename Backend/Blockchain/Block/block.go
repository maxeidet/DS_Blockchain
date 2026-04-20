package block

import (
	"crypto/sha256"
	"fmt"
	"time"
)

type Block struct {
	Index     int
	Timestamp int64
	Data      string
	PrevHash  string
	Hash      string
	Nonce     int
}

func NewBlock(index int, data, prevHash string) *Block {
	b := &Block{
		Index:     index,
		Timestamp: time.Now().Unix(),
		Data:      data,
		PrevHash:  prevHash,
	}
	b.Hash = b.CalculateHash()
	return b
}

func (b *Block) CalculateHash() string {
	record := fmt.Sprintf("%d%d%s%s%d", b.Index, b.Timestamp, b.Data, b.PrevHash, b.Nonce)
	h := sha256.Sum256([]byte(record))
	return fmt.Sprintf("%x", h)
}
