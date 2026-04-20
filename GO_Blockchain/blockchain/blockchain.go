package blockchain

import (
	"errors"
	"fmt"
	"strings"
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
		return fmt.Errorf("bara transfer-transaktioner får läggas i mempool, fick: %s", tx.Type)
	}

	if err := bc.ValidateTransaction(tx); err != nil {
		return err
	}

	added := bc.Mempool.Add(tx)
	if !added {
		return fmt.Errorf("transaktionen finns redan i mempool: %s", tx.ID)
	}

	return nil
}

func (bc *Blockchain) MineMempool(minerAddress string) (Block, error) {
	if minerAddress == "" {
		return Block{}, errors.New("miner address saknas")
	}

	mempoolTxs := bc.Mempool.GetAll()
	if len(mempoolTxs) == 0 {
		return Block{}, errors.New("mempool är tom")
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

	// Rensa mempool från transaktioner som redan finns i den nya kedjan
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
		return fmt.Errorf("fel index: got %d want %d", newBlock.Index, prevBlock.Index+1)
	}

	if newBlock.PrevHash != prevBlock.Hash {
		return errors.New("prev_hash matchar inte föregående block")
	}

	recalculatedHash := CalculateHash(newBlock)
	if newBlock.Hash != recalculatedHash {
		return errors.New("blockets hash är ogiltig")
	}

	targetPrefix := strings.Repeat("0", newBlock.Difficulty)
	if !strings.HasPrefix(newBlock.Hash, targetPrefix) {
		return fmt.Errorf("hash uppfyller inte difficulty-kravet: %s", newBlock.Hash)
	}

	return nil
}

func (bc *Blockchain) IsValid() error {
	return bc.ValidateChain(bc.Blocks)
}

func (bc *Blockchain) ValidateChain(blocks []Block) error {
	if len(blocks) == 0 {
		return errors.New("kedjan är tom")
	}

	genesis := blocks[0]
	if genesis.Index != 0 {
		return errors.New("genesis har fel index")
	}

	if genesis.Hash != CalculateHash(genesis) {
		return errors.New("genesis hash är ogiltig")
	}

	targetPrefix := strings.Repeat("0", genesis.Difficulty)
	if !strings.HasPrefix(genesis.Hash, targetPrefix) {
		return errors.New("genesis uppfyller inte difficulty")
	}

	if err := bc.ValidateGenesisTransactions(genesis.Transactions); err != nil {
		return fmt.Errorf("genesis-transaktioner är ogiltiga: %w", err)
	}

	for i := 1; i < len(blocks); i++ {
		if err := IsBlockValid(blocks[i], blocks[i-1]); err != nil {
			return fmt.Errorf("block %d ogiltigt: %w", i, err)
		}

		if err := bc.ValidateBlockTransactions(blocks[i].Transactions); err != nil {
			return fmt.Errorf("block %d har ogiltiga transaktioner: %w", i, err)
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

func (bc *Blockchain) ValidateTransaction(tx Transaction) error {
	if tx.To == "" {
		return errors.New("mottagaradress saknas")
	}

	if tx.Amount <= 0 {
		return errors.New("amount måste vara större än 0")
	}

	switch tx.Type {
	case "genesis":
		if tx.From != "" {
			return errors.New("genesis-transaktion får inte ha avsändare")
		}
		return nil

	case "coinbase":
		if tx.From != "" {
			return errors.New("coinbase-transaktion får inte ha avsändare")
		}
		if tx.Amount > bc.MiningReward {
			return fmt.Errorf("coinbase reward är för hög: %d > %d", tx.Amount, bc.MiningReward)
		}
		return nil

	case "transfer":
		if tx.From == "" {
			return errors.New("transfer-transaktion måste ha avsändare")
		}

		if tx.From == tx.To {
			return errors.New("kan inte skicka till samma adress")
		}

		balance := bc.GetBalance(tx.From)
		if balance < tx.Amount {
			return fmt.Errorf("otillräckligt saldo: balance=%d amount=%d", balance, tx.Amount)
		}

		return nil

	default:
		return fmt.Errorf("okänd transaktionstyp: %s", tx.Type)
	}
}

func (bc *Blockchain) ValidateGenesisTransactions(txs []Transaction) error {
	if len(txs) == 0 {
		return errors.New("genesis måste innehålla minst en transaktion")
	}

	seenTxIDs := make(map[string]bool)

	for _, tx := range txs {
		if seenTxIDs[tx.ID] {
			return fmt.Errorf("dublett-transaktion i genesis: %s", tx.ID)
		}
		seenTxIDs[tx.ID] = true

		if tx.Type != "genesis" {
			return fmt.Errorf("genesis-blocket får bara innehålla genesis-transaktioner, fick: %s", tx.Type)
		}

		if tx.From != "" {
			return errors.New("genesis-transaktioner måste ha tom avsändare")
		}

		if tx.To == "" {
			return errors.New("genesis-transaktion saknar mottagare")
		}

		if tx.Amount <= 0 {
			return fmt.Errorf("genesis-transaktion har ogiltigt amount: %d", tx.Amount)
		}
	}

	return nil
}

func (bc *Blockchain) ValidateBlockTransactions(txs []Transaction) error {
	if len(txs) == 0 {
		return errors.New("kan inte skapa block utan transaktioner")
	}

	spentInBlock := make(map[string]int)
	seenTxIDs := make(map[string]bool)
	coinbaseCount := 0

	for i, tx := range txs {
		if seenTxIDs[tx.ID] {
			return fmt.Errorf("dublett-transaktion i blocket: %s", tx.ID)
		}
		seenTxIDs[tx.ID] = true

		if tx.To == "" {
			return errors.New("mottagaradress saknas")
		}

		if tx.Amount <= 0 {
			return fmt.Errorf("ogiltigt amount i tx %s", tx.ID)
		}

		switch tx.Type {
		case "coinbase":
			coinbaseCount++

			if coinbaseCount > 1 {
				return errors.New("ett block får bara innehålla en coinbase-transaktion")
			}

			if i != 0 {
				return errors.New("coinbase-transaktionen måste ligga först i blocket")
			}

			if tx.From != "" {
				return errors.New("coinbase-transaktion får inte ha avsändare")
			}

			if tx.Amount > bc.MiningReward {
				return fmt.Errorf("coinbase reward är för hög: %d > %d", tx.Amount, bc.MiningReward)
			}

		case "transfer":
			if tx.From == "" {
				return fmt.Errorf("transfer-transaktion %s saknar avsändare", tx.ID)
			}

			if tx.From == tx.To {
				return fmt.Errorf("avsändare och mottagare är samma i tx %s", tx.ID)
			}

			spentInBlock[tx.From] += tx.Amount

		case "genesis":
			return errors.New("genesis-transaktioner får inte förekomma i vanliga block")

		default:
			return fmt.Errorf("okänd transaktionstyp: %s", tx.Type)
		}
	}

	for from, totalSpent := range spentInBlock {
		balance := bc.GetBalance(from)
		if totalSpent > balance {
			return fmt.Errorf("adress %s försöker spendera %d men har bara %d", from, totalSpent, balance)
		}
	}

	return nil
}