package blockchain

import "strings"

func MineBlock(block *Block) {
	targetPrefix := strings.Repeat("0", block.Difficulty)

	for {
		hash := CalculateHash(*block)
		if strings.HasPrefix(hash, targetPrefix) {
			block.Hash = hash
			return
		}
		block.Nonce++
	}
}