package blockchain

import "sync"

type Mempool struct {
	mu  sync.Mutex
	Txs map[string]Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		Txs: make(map[string]Transaction),
	}
}

func (mp *Mempool) Add(tx Transaction) bool {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if _, exists := mp.Txs[tx.ID]; exists {
		return false
	}

	mp.Txs[tx.ID] = tx
	return true
}

func (mp *Mempool) Remove(txID string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	delete(mp.Txs, txID)
}

func (mp *Mempool) RemoveMany(txIDs []string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	for _, txID := range txIDs {
		delete(mp.Txs, txID)
	}
}

func (mp *Mempool) GetAll() []Transaction {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	txs := make([]Transaction, 0, len(mp.Txs))
	for _, tx := range mp.Txs {
		txs = append(txs, tx)
	}

	return txs
}

func (mp *Mempool) Size() int {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	return len(mp.Txs)
}