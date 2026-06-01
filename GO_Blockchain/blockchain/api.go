package blockchain

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type API struct {
	BC     *Blockchain
	P2P    *P2PNode
	Wallet *Wallet
	Faucet *Wallet
}

func NewAPI(bc *Blockchain, p2p *P2PNode, wallet *Wallet, faucet *Wallet) *API {
	return &API{
		BC:     bc,
		P2P:    p2p,
		Wallet: wallet,
		Faucet: faucet,
	}
}

func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/status", api.handleStatus)
	mux.HandleFunc("/blocks", api.handleBlocks)
	mux.HandleFunc("/mempool", api.handleMempool)
	mux.HandleFunc("/transactions", api.handleTransactions)
	mux.HandleFunc("/transactions/manual", api.handleTransactionsManual)
	mux.HandleFunc("/faucet", api.handleFaucet)
	mux.HandleFunc("/mine", api.handleMine)
	mux.HandleFunc("/balance/", api.handleBalance)
	mux.HandleFunc("/peers", api.handlePeers)
	mux.HandleFunc("/attack-test", api.handleAttackTest)
}

func (api *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	walletAddress := ""
	walletNonce := uint64(0)
	if api.Wallet != nil {
		walletAddress = api.Wallet.Address
		walletNonce = api.BC.GetNonce(api.Wallet.Address)
	}

	faucetAddress := ""
	if api.Faucet != nil {
		faucetAddress = api.Faucet.Address
	}

	resp := map[string]any{
		"height":         len(api.BC.Blocks) - 1,
		"difficulty":     api.BC.Difficulty,
		"mining_reward":  api.BC.MiningReward,
		"mempool_size":   api.BC.Mempool.Size(),
		"is_valid":       api.BC.IsValid() == nil,
		"wallet_address": walletAddress,
		"wallet_nonce":   walletNonce,
		"faucet_address": faucetAddress,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handleBlocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, api.BC.Blocks)
}

func (api *API) handleMempool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, api.BC.Mempool.GetAll())
}

type createTransactionRequest struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

func (api *API) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if api.Wallet == nil {
		writeError(w, http.StatusInternalServerError, "wallet not initialized")
		return
	}

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	nonce := api.BC.NextNonce(api.Wallet.Address)

	tx, err := NewSignedTransferTransaction(api.Wallet, req.To, req.Amount, nonce)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.BC.AddTransactionToMempool(tx); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if api.P2P != nil {
		log.Printf("API announcing signed tx %s -> %s (%d) nonce=%d\n", tx.From, tx.To, tx.Amount, tx.Nonce)
		api.P2P.AnnounceTransaction(tx)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "signed transaction added to mempool",
		"tx":      tx,
	})
}

type faucetRequest struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

func (api *API) handleFaucet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if api.Faucet == nil {
		writeError(w, http.StatusInternalServerError, "faucet not initialized")
		return
	}

	var req faucetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	nonce := api.BC.NextNonce(api.Faucet.Address)

	tx, err := NewSignedTransferTransaction(
		api.Faucet,
		req.To,
		req.Amount,
		nonce,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.BC.AddTransactionToMempool(tx); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if api.P2P != nil {
		log.Printf(
			"API announcing faucet tx %s -> %s (%d) nonce=%d\n",
			tx.From,
			tx.To,
			tx.Amount,
			tx.Nonce,
		)
		api.P2P.AnnounceTransaction(tx)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "faucet transaction added to mempool",
		"tx":      tx,
	})
}

type mineRequest struct {
	MinerAddress string `json:"miner_address"`
}

func (api *API) handleMine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req mineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	block, err := api.BC.MineMempool(req.MinerAddress)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if api.P2P != nil {
		log.Printf("API announcing block index=%d hash=%s\n", block.Index, block.Hash)
		api.P2P.AnnounceBlock(block)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "block mined",
		"block":   block,
	})
}

func (api *API) handleBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	prefix := "/balance/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeError(w, http.StatusBadRequest, "invalid balance path")
		return
	}

	address := strings.TrimPrefix(r.URL.Path, prefix)
	if address == "" {
		writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	resp := map[string]any{
		"address": address,
		"balance": api.BC.GetBalance(address),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (api *API) handlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if api.P2P == nil {
		writeJSON(w, http.StatusOK, []string{})
		return
	}

	api.P2P.mu.Lock()
	peers := make([]string, 0, len(api.P2P.Peers))
	for peer := range api.P2P.Peers {
		peers = append(peers, peer)
	}
	api.P2P.mu.Unlock()

	writeJSON(w, http.StatusOK, peers)
}

// This is for testing nonce failure
type createTransactionWithNonceRequest struct {
	To     string `json:"to"`
	Amount int    `json:"amount"`
	Nonce  uint64 `json:"nonce"`
}
func (api *API) handleTransactionsManual(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if api.Wallet == nil {
		writeError(w, http.StatusInternalServerError, "wallet not initialized")
		return
	}

	var req createTransactionWithNonceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tx, err := NewSignedTransferTransaction(api.Wallet, req.To, req.Amount, req.Nonce)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := api.BC.AddTransactionToMempool(tx); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if api.P2P != nil {
		log.Printf("API announcing manual tx %s -> %s (%d) nonce=%d\n", tx.From, tx.To, tx.Amount, tx.Nonce)
		api.P2P.AnnounceTransaction(tx)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "manual signed transaction added to mempool",
		"tx":      tx,
	})
}




func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}