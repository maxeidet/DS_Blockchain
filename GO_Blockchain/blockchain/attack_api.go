package blockchain

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

// AttackResult describes the outcome of one attack attempt.
type AttackResult struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Blocked     bool   `json:"blocked"`  // true = system correctly rejected the attack
	Error       string `json:"error"`    // the rejection message from the blockchain
	Expected    string `json:"expected"` // what we expected to happen
}

// handleAttackTest runs all four Exp-4 attacks against a fresh, isolated
// blockchain instance (not the live node chain) and returns the results.
//
// POST /attack-test
func (api *API) handleAttackTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	results := runAttackSuite()
	writeJSON(w, http.StatusOK, map[string]any{
		"attacks": results,
		"summary": summarize(results),
	})
}

// runAttackSuite sets up an isolated blockchain and runs every attack.
func runAttackSuite() []AttackResult {
	// ── Setup: isolated blockchain with two wallets ───────────────────────────
	params := DefaultNetworkParams()
	bc := NewBlockchain(params)

	alice, _ := NewWallet()
	bob, _ := NewWallet()
	eve, _ := NewWallet()

	// Fund Alice via faucet so she has coins to spend.
	faucet, _ := FaucetWallet()
	fundTx, _ := NewSignedTransferTransaction(faucet, alice.Address, 200, 1)
	_ = bc.AddTransactionToMempool(fundTx)
	_, _ = bc.MineMempool(bob.Address) // mine the funding block

	results := []AttackResult{}
	results = append(results, attackReplay(bc, alice, bob))
	results = append(results, attackInvalidSignature(bc, alice, bob))
	results = append(results, attackWrongPublicKey(bc, alice, bob, eve))
	results = append(results, attackOverdraft(bc, alice, bob))
	results = append(results, attackFutureTimestamp(bc, alice, bob))
	return results
}

// ── Attack A: Replay ──────────────────────────────────────────────────────────
// Send the exact same signed transaction twice. The second attempt must fail.
func attackReplay(bc *Blockchain, alice, bob *Wallet) AttackResult {
	nonce := bc.NextNonce(alice.Address)
	tx, err := NewSignedTransferTransaction(alice, bob.Address, 10, nonce)
	if err != nil {
		return AttackResult{Name: "Replay Attack", Error: "setup error: " + err.Error()}
	}

	// First submission – should succeed.
	if err := bc.AddTransactionToMempool(tx); err != nil {
		return AttackResult{
			Name:        "Replay Attack",
			Description: "Submit same signed transaction twice. Second attempt must be rejected.",
			Blocked:     false,
			Error:       "first submission already failed: " + err.Error(),
			Expected:    "Second submission rejected (nonce or duplicate ID)",
		}
	}

	// Replay: submit exactly the same transaction object again.
	replayErr := bc.AddTransactionToMempool(tx)

	blocked := replayErr != nil
	errMsg := ""
	if replayErr != nil {
		errMsg = replayErr.Error()
	}

	// Clean up mempool for subsequent attacks.
	bc.Mempool.Remove(tx.ID)

	return AttackResult{
		Name:        "Replay Attack",
		Description: "Submit same signed transaction twice. Second attempt must be rejected.",
		Blocked:     blocked,
		Error:       errMsg,
		Expected:    "Rejected: duplicate tx ID or nonce already used",
	}
}

// ── Attack B: Invalid Signature ───────────────────────────────────────────────
// Replace the signature with random bytes; the ECDSA check must catch this.
func attackInvalidSignature(bc *Blockchain, alice, bob *Wallet) AttackResult {
	nonce := bc.NextNonce(alice.Address)
	tx, err := NewSignedTransferTransaction(alice, bob.Address, 10, nonce)
	if err != nil {
		return AttackResult{Name: "Invalid Signature", Error: "setup error: " + err.Error()}
	}

	// Corrupt the signature (flip every byte).
	sigBytes, _ := hex.DecodeString(tx.Signature)
	for i := range sigBytes {
		sigBytes[i] ^= 0xFF
	}
	tx.Signature = hex.EncodeToString(sigBytes)
	// Recalculate ID so it doesn't look like a duplicate (signature is part of ID hash).
	tx.ID = CalculateTransactionID(tx)

	rejectErr := bc.AddTransactionToMempool(tx)
	blocked := rejectErr != nil
	errMsg := ""
	if rejectErr != nil {
		errMsg = rejectErr.Error()
	}

	return AttackResult{
		Name:        "Invalid Signature",
		Description: "Flip all bits in the ECDSA signature. ECDSA verify must reject it.",
		Blocked:     blocked,
		Error:       errMsg,
		Expected:    "Rejected: invalid signature",
	}
}

// ── Attack C: Wrong Public Key ─────────────────────────────────────────────────
// Use Alice's private key to sign but claim Eve's public key (and address).
// The derived address from Eve's pubkey won't match the signature made by Alice.
func attackWrongPublicKey(bc *Blockchain, alice, bob, eve *Wallet) AttackResult {
	nonce := bc.NextNonce(alice.Address)

	// Build the transaction as if it comes from Eve, but sign with Alice's key.
	tx := Transaction{
		From:      eve.Address,
		To:        bob.Address,
		Amount:    10,
		Timestamp: time.Now().UnixMilli(),
		Type:      "transfer",
		Nonce:     nonce,
		PublicKey: hex.EncodeToString(eve.PublicKey), // Eve's pubkey
	}

	signBytes, _ := tx.SigningBytes()
	sig, _ := SignBytes(alice.PrivateKey, signBytes) // signed by Alice!
	tx.Signature = sig
	tx.ID = CalculateTransactionID(tx)

	rejectErr := bc.AddTransactionToMempool(tx)
	blocked := rejectErr != nil
	errMsg := ""
	if rejectErr != nil {
		errMsg = rejectErr.Error()
	}

	return AttackResult{
		Name:        "Wrong Public Key",
		Description: "Sign with Alice's private key but claim Eve's public key/address. Address derivation check must catch the mismatch.",
		Blocked:     blocked,
		Error:       errMsg,
		Expected:    "Rejected: invalid signature or address mismatch",
	}
}

// ── Attack D: Overdraft ────────────────────────────────────────────────────────
// Try to send more coins than Alice's balance.
func attackOverdraft(bc *Blockchain, alice, bob *Wallet) AttackResult {
	balance := bc.GetBalance(alice.Address)
	nonce := bc.NextNonce(alice.Address)

	overAmount := balance + 9999
	tx, err := NewSignedTransferTransaction(alice, bob.Address, overAmount, nonce)
	if err != nil {
		return AttackResult{Name: "Overdraft", Error: "setup error: " + err.Error()}
	}

	rejectErr := bc.AddTransactionToMempool(tx)
	blocked := rejectErr != nil
	errMsg := ""
	if rejectErr != nil {
		errMsg = rejectErr.Error()
	}

	return AttackResult{
		Name: "Overdraft",
		Description: fmt.Sprintf(
			"Attempt to spend %d coins when Alice only has %d. Balance check must reject.",
			overAmount, balance,
		),
		Blocked:  blocked,
		Error:    errMsg,
		Expected: "Rejected: insufficient balance",
	}
}

// ── Attack E: Future Timestamp ────────────────────────────────────────────────
// Mine a block with a timestamp 3 hours in the future.
// The timestamp validation added to IsBlockValid must reject it.
func attackFutureTimestamp(bc *Blockchain, alice, bob *Wallet) AttackResult {
	// Build a minimal valid block but set timestamp far into the future.
	prev := bc.LatestBlock()
	coinbase := NewCoinbaseTransaction(bob.Address, bc.MiningReward)

	badBlock := Block{
		Index:        prev.Index + 1,
		Timestamp:    time.Now().Add(3 * time.Hour).UnixMilli(), // 3 hours in the future
		Transactions: []Transaction{coinbase},
		PrevHash:     prev.Hash,
		Nonce:        0,
		Difficulty:   bc.Difficulty,
	}
	MineBlock(&badBlock)

	rejectErr := IsBlockValid(badBlock, prev)
	blocked := rejectErr != nil
	errMsg := ""
	if rejectErr != nil {
		errMsg = rejectErr.Error()
	}

	return AttackResult{
		Name:        "Future Timestamp",
		Description: "Mine a block with a timestamp 3 hours in the future. Timestamp validation must reject it.",
		Blocked:     blocked,
		Error:       errMsg,
		Expected:    "Rejected: block timestamp is too far in the future",
	}
}

// ── Helper ────────────────────────────────────────────────────────────────────
func summarize(results []AttackResult) map[string]int {
	blocked, passed := 0, 0
	for _, r := range results {
		if r.Blocked {
			blocked++
		} else {
			passed++
		}
	}
	return map[string]int{
		"total":   len(results),
		"blocked": blocked,
		"leaked":  passed,
	}
}
