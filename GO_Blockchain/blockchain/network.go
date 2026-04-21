package blockchain

import "sort"

type NetworkParams struct {
	NetworkID        string
	Difficulty       int
	MiningReward     int
	GenesisTimestamp int64
	GenesisBalances  map[string]int
}

func DefaultNetworkParams() NetworkParams {
	faucet, err := FaucetWallet()
	if err != nil {
		panic(err)
	}

	return NetworkParams{
		NetworkID:    "go-blockchain-demo",
		Difficulty:   3,
		MiningReward: 50,
		GenesisBalances: map[string]int{
			faucet.Address: 1000,
			"addr_alice":   100,
			"addr_bob":     50,
			"addr_charlie": 25,
		},
		GenesisTimestamp: 1700000000000,
	}
}

func BuildGenesisTransactions(params NetworkParams) []Transaction {
	txs := make([]Transaction, 0, len(params.GenesisBalances))

	// Viktigt: ordningen måste vara deterministisk på alla noder
	addresses := make([]string, 0, len(params.GenesisBalances))
	for addr := range params.GenesisBalances {
		addresses = append(addresses, addr)
	}
	sort.Strings(addresses)

	for _, addr := range addresses {
		amount := params.GenesisBalances[addr]
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