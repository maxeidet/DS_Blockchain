package blockchain

import (
	"ds-project/Blockchain/Block"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type Blockchain struct {
	Chain []*block.Block
	Nodes map[string]bool
}

func NewBlockchain() *Blockchain {
	genesis := block.NewBlock(0, "Genesis Block", "0")
	genesis.Mine()
	return &Blockchain{
		Chain: []*block.Block{genesis},
		Nodes: make(map[string]bool),
	}
}

func (bc *Blockchain) AddBlock(data string) *block.Block {
	prev := bc.Chain[len(bc.Chain)-1]
	b := block.NewBlock(len(bc.Chain), data, prev.Hash)
	b.Mine()
	bc.Chain = append(bc.Chain, b)
	return b
}

func (bc *Blockchain) validateChain(chain []*block.Block) bool {
	target := strings.Repeat("0", block.Difficulty)
	for i := 1; i < len(chain); i++ {
		curr := chain[i]
		prev := chain[i-1]
		if curr.Hash != curr.CalculateHash() {
			return false
		}
		if curr.PrevHash != prev.Hash {
			return false
		}
		if !strings.HasPrefix(curr.Hash, target) {
			return false
		}
	}
	return true
}

func (bc *Blockchain) IsValid() bool {
	return bc.validateChain(bc.Chain)
}

func (bc *Blockchain) ResolveConflicts() bool {
	var newChain []*block.Block
	maxLength := len(bc.Chain)

	for node := range bc.Nodes {
		resp, err := http.Get(node + "/chain")
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var chain []*block.Block
			if err := json.NewDecoder(resp.Body).Decode(&chain); err == nil {
				if len(chain) > maxLength && bc.validateChain(chain) {
					maxLength = len(chain)
					newChain = chain
				}
			}
		}
		resp.Body.Close()
	}

	if newChain != nil {
		bc.Chain = newChain
		return true
	}

	return false
}

func (bc *Blockchain) RegisterNode(address string) {
	if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
		address = "http://" + address
	}
	u, err := url.Parse(address)
	if err == nil && u.Host != "" {
		bc.Nodes[u.Scheme+"://"+u.Host] = true
	}
}
