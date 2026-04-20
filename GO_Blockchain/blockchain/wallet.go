package blockchain

import (
	"crypto/rand"
	"encoding/hex"
)

type Wallet struct {
	Address string `json:"address"`
}

func NewWallet() Wallet {
	b := make([]byte, 8)
	_, _ = rand.Read(b)

	return Wallet{
		Address: "addr_" + hex.EncodeToString(b),
	}
}