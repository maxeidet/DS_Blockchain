package blockchain

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Blockchain struct {
	Blocks       []Block
	Mempool      *Mempool
	Difficulty   int
	MiningReward int
	NetworkID    string
}

func NewBlockchain(params NetworkParams) *Blockchain {
	genesis := BuildGenesisBlock(params)

	return &Blockchain{
		Blocks:       []Block{genesis},
		Mempool:      NewMempool(),
		Difficulty:   params.Difficulty,
		MiningReward: params.MiningReward,
		NetworkID:    params.NetworkID,
	}
}

func (bc *Blockchain) LatestBlock() Block {
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) Height() int {
	return len(bc.Blocks) - 1
}

func (bc *Blockchain) GenesisHash() string {
	if len(bc.Blocks) == 0 {
		return ""
	}
	return bc.Blocks[0].Hash
}

func (bc *Blockchain) BestHash() string {
	if len(bc.Blocks) == 0 {
		return ""
	}
	return bc.LatestBlock().Hash
}

func (bc *Blockchain) AddBlock(transactions []Transaction) (Block, error) {
	prevBlock := bc.LatestBlock()

	if err := bc.ValidateBlockTransactions(transactions); err != nil {
		return Block{}, err
	}

	newBlock := NewBlock(
		prevBlock.Index+1,
		transactions,
		prevBlock.Hash,
		bc.Difficulty,
	)

	MineBlock(&newBlock)

	if err := IsBlockValid(newBlock, prevBlock); err != nil {
		return Block{}, err
	}

	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock, nil
}

func (bc *Blockchain) AddTransactionToMempool(tx Transaction) error {
	if tx.Type != "transfer" {
		return fmt.Errorf("only transfer transactions are allowed in mempool, got: %s", tx.Type)
	}

	if err := bc.ValidateTransaction(tx); err != nil {
		return err
	}

	added := bc.Mempool.Add(tx)
	if !added {
		return fmt.Errorf("transaction already exists in mempool: %s", tx.ID)
	}

	return nil
}

func (bc *Blockchain) MineMempool(minerAddress string) (Block, error) {
	if minerAddress == "" {
		return Block{}, errors.New("miner address missing")
	}

	mempoolTxs := bc.Mempool.GetAll()
	if len(mempoolTxs) == 0 {
		return Block{}, errors.New("mempool is empty")
	}

	coinbaseTx := NewCoinbaseTransaction(minerAddress, bc.MiningReward)

	allTxs := make([]Transaction, 0, len(mempoolTxs)+1)
	allTxs = append(allTxs, coinbaseTx)
	allTxs = append(allTxs, mempoolTxs...)

	block, err := bc.AddBlock(allTxs)
	if err != nil {
		return Block{}, err
	}

	txIDs := make([]string, 0, len(mempoolTxs))
	for _, tx := range mempoolTxs {
		txIDs = append(txIDs, tx.ID)
	}
	bc.Mempool.RemoveMany(txIDs)

	return block, nil
}

func (bc *Blockchain) ReplaceChainIfValid(newBlocks []Block) error {
	if len(newBlocks) == 0 {
		return errors.New("new chain is empty")
	}

	if len(newBlocks) <= len(bc.Blocks) {
		return errors.New("new chain is not longer than current chain")
	}

	if err := bc.ValidateChain(newBlocks); err != nil {
		return err
	}

	if len(bc.Blocks) > 0 && newBlocks[0].Hash != bc.Blocks[0].Hash {
		return errors.New("genesis mismatch")
	}

	bc.Blocks = cloneBlocks(newBlocks)

	chainTxIDs := make(map[string]bool)
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			chainTxIDs[tx.ID] = true
		}
	}

	for _, tx := range bc.Mempool.GetAll() {
		if chainTxIDs[tx.ID] {
			bc.Mempool.Remove(tx.ID)
		}
	}

	return nil
}

func cloneBlocks(src []Block) []Block {
	out := make([]Block, len(src))
	for i, b := range src {
		txCopy := make([]Transaction, len(b.Transactions))
		copy(txCopy, b.Transactions)

		out[i] = Block{
			Index:        b.Index,
			Timestamp:    b.Timestamp,
			Transactions: txCopy,
			PrevHash:     b.PrevHash,
			Hash:         b.Hash,
			Nonce:        b.Nonce,
			Difficulty:   b.Difficulty,
		}
	}
	return out
}

func IsBlockValid(newBlock, prevBlock Block) error {
	if newBlock.Index != prevBlock.Index+1 {
		return fmt.Errorf("invalid index: got %d want %d", newBlock.Index, prevBlock.Index+1)
	}

	if newBlock.PrevHash != prevBlock.Hash {
		return errors.New("prev_hash does not match previous block")
	}

	recalculatedHash := CalculateHash(newBlock)
	if newBlock.Hash != recalculatedHash {
		return errors.New("invalid block hash")
	}

	targetPrefix := strings.Repeat("0", newBlock.Difficulty)
	if !strings.HasPrefix(newBlock.Hash, targetPrefix) {
		return fmt.Errorf("hash does not meet difficulty requirement: %s", newBlock.Hash)
	}
	if newBlock.Timestamp > time.Now().Add(2*time.Hour).UnixMilli() {
		return errors.New("block timestamp is too far in the future")
	}


	return nil
}

func (bc *Blockchain) IsValid() error {
	return bc.ValidateChain(bc.Blocks)
}

func (bc *Blockchain) ValidateChain(blocks []Block) error {
	if len(blocks) == 0 {
		return errors.New("chain is empty")
	}

	genesis := blocks[0]
	if genesis.Index != 0 {
		return errors.New("genesis has wrong index")
	}

	if genesis.Hash != CalculateHash(genesis) {
		return errors.New("invalid genesis hash")
	}

	targetPrefix := strings.Repeat("0", genesis.Difficulty)
	if !strings.HasPrefix(genesis.Hash, targetPrefix) {
		return errors.New("genesis does not meet difficulty")
	}

	if err := bc.ValidateGenesisTransactions(genesis.Transactions); err != nil {
		return fmt.Errorf("invalid genesis transactions: %w", err)
	}

	balances := make(map[string]int)
	nonces := make(map[string]uint64)
	seenTxIDs := make(map[string]bool)

	for _, tx := range genesis.Transactions {
		if seenTxIDs[tx.ID] {
			return fmt.Errorf("duplicate transaction in genesis: %s", tx.ID)
		}
		seenTxIDs[tx.ID] = true
		applyTransactionToState(tx, balances, nonces)
	}

	for i := 1; i < len(blocks); i++ {
		curr := blocks[i]
		prev := blocks[i-1]

		if err := IsBlockValid(curr, prev); err != nil {
			return fmt.Errorf("invalid block %d: %w", i, err)
		}

		if err := bc.ValidateBlockTransactionsAgainstState(curr.Transactions, balances, nonces, seenTxIDs); err != nil {
			return fmt.Errorf("block %d has invalid transactions: %w", i, err)
		}

		for _, tx := range curr.Transactions {
			seenTxIDs[tx.ID] = true
			applyTransactionToState(tx, balances, nonces)
		}
	}

	return nil
}

func (bc *Blockchain) GetBalance(address string) int {
	balance := 0

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.From == address {
				balance -= tx.Amount
			}
			if tx.To == address {
				balance += tx.Amount
			}
		}
	}

	return balance
}

func (bc *Blockchain) GetNonce(address string) uint64 {
	var nonce uint64 = 0

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Type == "transfer" && tx.From == address {
				if tx.Nonce > nonce {
					nonce = tx.Nonce
				}
			}
		}
	}

	return nonce
}

func (bc *Blockchain) NextNonce(address string) uint64 {
	maxNonce := bc.GetNonce(address)

	for _, tx := range bc.Mempool.GetAll() {
		if tx.Type == "transfer" && tx.From == address {
			if tx.Nonce > maxNonce {
				maxNonce = tx.Nonce
			}
		}
	}

	return maxNonce + 1
}

func (bc *Blockchain) ValidateTransaction(tx Transaction) error {
	if tx.To == "" {
		return errors.New("recipient address missing")
	}

	if tx.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}

	switch tx.Type {
	case "genesis":
		if tx.From != "" {
			return errors.New("genesis transaction cannot have sender")
		}
		if tx.Nonce != 0 {
			return errors.New("genesis transaction cannot have nonce")
		}
		return nil

	case "coinbase":
		if tx.From != "" {
			return errors.New("coinbase transaction cannot have sender")
		}
		if tx.Amount > bc.MiningReward {
			return fmt.Errorf("coinbase reward is too high: %d > %d", tx.Amount, bc.MiningReward)
		}
		if tx.Nonce != 0 {
			return errors.New("coinbase transaction cannot have nonce")
		}
		return nil

	case "transfer":
		if err := bc.validateTransferBasics(tx); err != nil {
			return err
		}

		expectedNonce := bc.NextNonce(tx.From)
		if tx.Nonce != expectedNonce {
			return fmt.Errorf("invalid nonce: got %d want %d", tx.Nonce, expectedNonce)
		}

		balance := bc.GetBalance(tx.From)
		if balance < tx.Amount {
			return fmt.Errorf("insufficient balance: balance=%d amount=%d", balance, tx.Amount)
		}

		return nil

	default:
		return fmt.Errorf("unknown transaction type: %s", tx.Type)
	}
}

func (bc *Blockchain) ValidateGenesisTransactions(txs []Transaction) error {
	if len(txs) == 0 {
		return errors.New("genesis must contain at least one transaction")
	}

	seenTxIDs := make(map[string]bool)

	for _, tx := range txs {
		if seenTxIDs[tx.ID] {
			return fmt.Errorf("duplicate transaction in genesis: %s", tx.ID)
		}
		seenTxIDs[tx.ID] = true

		if tx.Type != "genesis" {
			return fmt.Errorf("genesis block can only contain genesis transactions, got: %s", tx.Type)
		}

		if tx.From != "" {
			return errors.New("genesis transactions must have empty sender")
		}

		if tx.To == "" {
			return errors.New("genesis transaction missing recipient")
		}

		if tx.Amount <= 0 {
			return fmt.Errorf("genesis transaction has invalid amount: %d", tx.Amount)
		}

		if tx.Nonce != 0 {
			return errors.New("genesis transaction cannot have nonce")
		}
	}

	return nil
}

func (bc *Blockchain) ValidateBlockTransactions(txs []Transaction) error {
	if len(txs) == 0 {
		return errors.New("cannot create block without transactions")
	}

	balances := bc.buildBalancesFromCurrentChain()
	nonces := bc.buildNoncesFromCurrentChain()
	seenTxIDs := bc.buildSeenTxIDsFromCurrentChain()

	return bc.ValidateBlockTransactionsAgainstState(txs, balances, nonces, seenTxIDs)
}

func (bc *Blockchain) ValidateBlockTransactionsAgainstState(
	txs []Transaction,
	balances map[string]int,
	nonces map[string]uint64,
	seenTxIDs map[string]bool,
) error {
	if len(txs) == 0 {
		return errors.New("cannot create block without transactions")
	}

	tempBalances := copyBalances(balances)
	tempNonces := copyNonces(nonces)
	coinbaseCount := 0

	for i, tx := range txs {
		if seenTxIDs[tx.ID] {
			return fmt.Errorf("transaction already exists in chain: %s", tx.ID)
		}

		if tx.To == "" {
			return errors.New("recipient address missing")
		}

		if tx.Amount <= 0 {
			return fmt.Errorf("invalid amount in tx %s", tx.ID)
		}

		switch tx.Type {
		case "coinbase":
			coinbaseCount++

			if coinbaseCount > 1 {
				return errors.New("a block can only contain one coinbase transaction")
			}

			if i != 0 {
				return errors.New("coinbase transaction must be first in block")
			}

			if tx.From != "" {
				return errors.New("coinbase transaction cannot have sender")
			}

			if tx.Amount > bc.MiningReward {
				return fmt.Errorf("coinbase reward is too high: %d > %d", tx.Amount, bc.MiningReward)
			}

			if tx.Nonce != 0 {
				return errors.New("coinbase transaction cannot have nonce")
			}

			applyTransactionToState(tx, tempBalances, tempNonces)

		case "transfer":
			if tx.From == "" {
				return fmt.Errorf("transfer transaction %s missing sender", tx.ID)
			}

			if err := bc.validateTransferBasics(tx); err != nil {
				return fmt.Errorf("invalid transfer transaction %s: %w", tx.ID, err)
			}

			expectedNonce := tempNonces[tx.From] + 1
			if tx.Nonce != expectedNonce {
				return fmt.Errorf(
					"adress %s har invalid nonce: got %d want %d",
					tx.From,
					tx.Nonce,
					expectedNonce,
				)
			}

			if tempBalances[tx.From] < tx.Amount {
				return fmt.Errorf(
					"address %s trying to spend %d but only has %d",
					tx.From,
					tx.Amount,
					tempBalances[tx.From],
				)
			}

			applyTransactionToState(tx, tempBalances, tempNonces)

		case "genesis":
			return errors.New("genesis transactions cannot appear in regular blocks")

		default:
			return fmt.Errorf("unknown transaction type: %s", tx.Type)
		}
	}

	for addr, balance := range tempBalances {
		balances[addr] = balance
	}
	for addr, nonce := range tempNonces {
		nonces[addr] = nonce
	}

	return nil
}

func (bc *Blockchain) validateTransferBasics(tx Transaction) error {
	if tx.From == "" {
		return errors.New("transfer transaction must have sender")
	}

	if tx.From == tx.To {
		return errors.New("cannot send to the same address")
	}

	if tx.Nonce == 0 {
		return errors.New("transfer transaction must have nonce > 0")
	}

	if tx.PublicKey == "" {
		return errors.New("transfer transaction missing public key")
	}

	if tx.Signature == "" {
		return errors.New("transfer transaction missing signature")
	}

	pubKeyBytes, err := hex.DecodeString(tx.PublicKey)
	if err != nil {
		return errors.New("invalid public key encoding")
	}

	derivedAddress := AddressFromPublicKey(pubKeyBytes)
	if derivedAddress != tx.From {
		return errors.New("from address does not match public key")
	}

	signBytes, err := tx.SigningBytes()
	if err != nil {
		return fmt.Errorf("could not build signing bytes: %w", err)
	}

	ok, err := VerifySignature(pubKeyBytes, signBytes, tx.Signature)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}
	if !ok {
		return errors.New("invalid signature")
	}

	return nil
}

func (bc *Blockchain) buildBalancesFromCurrentChain() map[string]int {
	balances := make(map[string]int)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			applyTransactionToState(tx, balances, make(map[string]uint64))
		}
	}

	return balances
}

func (bc *Blockchain) buildNoncesFromCurrentChain() map[string]uint64 {
	nonces := make(map[string]uint64)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Type == "transfer" {
				if tx.Nonce > nonces[tx.From] {
					nonces[tx.From] = tx.Nonce
				}
			}
		}
	}

	return nonces
}

func (bc *Blockchain) buildSeenTxIDsFromCurrentChain() map[string]bool {
	seen := make(map[string]bool)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			seen[tx.ID] = true
		}
	}

	return seen
}

func copyBalances(src map[string]int) map[string]int {
	dst := make(map[string]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func copyNonces(src map[string]uint64) map[string]uint64 {
	dst := make(map[string]uint64, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func applyTransactionToState(tx Transaction, balances map[string]int, nonces map[string]uint64) {
	switch tx.Type {
	case "genesis", "coinbase":
		balances[tx.To] += tx.Amount
	case "transfer":
		balances[tx.From] -= tx.Amount
		balances[tx.To] += tx.Amount
		nonces[tx.From] = tx.Nonce
	}
}