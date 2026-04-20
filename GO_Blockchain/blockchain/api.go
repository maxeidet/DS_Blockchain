package blockchain

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type API struct {
	BC  *Blockchain
	P2P *P2PNode
}

func NewAPI(bc *Blockchain, p2p *P2PNode) *API {
	return &API{
		BC:  bc,
		P2P: p2p,
	}
}

func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/status", api.handleStatus)
	mux.HandleFunc("/blocks", api.handleBlocks)
	mux.HandleFunc("/mempool", api.handleMempool)
	mux.HandleFunc("/transactions", api.handleTransactions)
	mux.HandleFunc("/mine", api.handleMine)
	mux.HandleFunc("/balance/", api.handleBalance)
	mux.HandleFunc("/peers", api.handlePeers)
}

func (api *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	resp := map[string]any{
		"height":        len(api.BC.Blocks) - 1,
		"difficulty":    api.BC.Difficulty,
		"mining_reward": api.BC.MiningReward,
		"mempool_size":  api.BC.Mempool.Size(),
		"is_valid":      api.BC.IsValid() == nil,
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
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

func (api *API) handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tx := NewTransferTransaction(req.From, req.To, req.Amount)

	if err := api.BC.AddTransactionToMempool(tx); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if api.P2P != nil {
		log.Printf("API announcing tx %s -> %s (%d)\n", tx.From, tx.To, tx.Amount)
		api.P2P.AnnounceTransaction(tx)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "transaction added to mempool",
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