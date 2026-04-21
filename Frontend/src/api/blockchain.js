/**
 * Blockchain Node API Client — GO_Blockchain
 *
 * Matches the HTTP API exposed by GO_Blockchain/main.go:
 *
 *   GET  /status
 *   GET  /blocks
 *   GET  /mempool
 *   GET  /balance/{address}
 *   GET  /peers
 *   POST /transactions        { to, amount }
 *   POST /faucet              { to, amount }
 *   POST /mine                { miner_address }
 */

const DEFAULT_TIMEOUT_MS = 5000;

async function fetchWithTimeout(url, options = {}, timeoutMs = DEFAULT_TIMEOUT_MS) {
  const controller = new AbortController();
  const id = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const res = await fetch(url, { ...options, signal: controller.signal });
    clearTimeout(id);
    if (!res.ok) throw new Error(`HTTP ${res.status}: ${res.statusText}`);
    return await res.json();
  } catch (err) {
    clearTimeout(id);
    throw err;
  }
}

// ─── Status ───────────────────────────────────────────────────────────────────

/**
 * GET /status
 * Returns: { height, difficulty, mining_reward, mempool_size, is_valid,
 *             wallet_address, wallet_nonce, faucet_address }
 */
export async function getStatus(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/status`);
}

// ─── Chain ────────────────────────────────────────────────────────────────────

/**
 * GET /blocks
 * Returns: Block[] — array of all blocks in order
 */
export async function getBlocks(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/blocks`);
}

// ─── Mempool ──────────────────────────────────────────────────────────────────

/**
 * GET /mempool
 * Returns: Transaction[] — pending transactions
 */
export async function getMempool(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/mempool`);
}

// ─── Peers ────────────────────────────────────────────────────────────────────

/**
 * GET /peers
 * Returns: string[] — list of peer addresses
 */
export async function getPeers(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/peers`);
}

// ─── Balance ──────────────────────────────────────────────────────────────────

/**
 * GET /balance/{address}
 * Returns: { address, balance }
 */
export async function getBalance(nodeUrl, address) {
  return fetchWithTimeout(`${nodeUrl}/balance/${encodeURIComponent(address)}`);
}

// ─── Transactions ─────────────────────────────────────────────────────────────

/**
 * POST /transactions
 * Body: { to: string, amount: number }
 * Returns: { message, tx }
 *
 * Note: The node signs the transaction automatically with its own wallet.
 */
export async function submitTransaction(nodeUrl, { to, amount }) {
  return fetchWithTimeout(`${nodeUrl}/transactions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ to, amount: Number(amount) }),
  });
}

// ─── Faucet ───────────────────────────────────────────────────────────────────

/**
 * POST /faucet
 * Body: { to: string, amount: number }
 * Returns: { message, tx }
 */
export async function faucetDrop(nodeUrl, { to, amount }) {
  return fetchWithTimeout(`${nodeUrl}/faucet`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ to, amount: Number(amount) }),
  });
}

// ─── Mining ───────────────────────────────────────────────────────────────────

/**
 * POST /mine
 * Body: { miner_address: string }
 * Returns: { message, block }
 */
export async function mineBlock(nodeUrl, minerAddress = '') {
  return fetchWithTimeout(`${nodeUrl}/mine`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ miner_address: minerAddress }),
  });
}
