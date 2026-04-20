package blockchain

type NetworkParams struct {
	NetworkID        string
	Difficulty       int
	MiningReward     int
	GenesisTimestamp int64
	GenesisBalances  map[string]int
}

func DefaultNetworkParams() NetworkParams {
	return NetworkParams{
		NetworkID:        "go-blockchain-demo",
		Difficulty:       3,
		MiningReward:     50,
		GenesisTimestamp: 1700000000000,
		GenesisBalances: map[string]int{
			"addr_alice":   100,
			"addr_bob":     50,
			"addr_charlie": 25,
		},
	}
}

func BuildGenesisTransactions(params NetworkParams) []Transaction {
	txs := make([]Transaction, 0, len(params.GenesisBalances))

	// Viktigt: ordningen måste vara deterministisk
	addresses := []string{"addr_alice", "addr_bob", "addr_charlie"}

	for _, addr := range addresses {
		amount, ok := params.GenesisBalances[addr]
		if !ok {
			continue
		}

		txs = append(txs, NewGenesisTransactionWithTimestamp(addr, amount, params.GenesisTimestamp))
	}

	return txs
}

func BuildGenesisBlock(params NetworkParams) Block {
	txs := BuildGenesisTransactions(params)

	block := Block{
		Index:        0,
		Timestamp:    params.GenesisTimestamp,
		Transactions: txs,
		PrevHash:     "",
		Nonce:        0,
		Difficulty:   params.Difficulty,
	}

	MineBlock(&block)
	return block
}