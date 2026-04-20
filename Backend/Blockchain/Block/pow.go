package block

import "strings"

const Difficulty = 3 // hash must start with "000"

func (b *Block) Mine() {
	target := strings.Repeat("0", Difficulty)
	for !strings.HasPrefix(b.Hash, target) {
		b.Nonce++
		b.Hash = b.CalculateHash()
	}
}
