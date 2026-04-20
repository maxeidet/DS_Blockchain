/**
 * Blockchain Node API Client
 *
 * Each node exposes an HTTP API on localhost at a given port.
 * This module provides functions to query any node by its base URL.
 *
 * Adjust endpoint paths to match your actual backend routes.
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

// ─── Chain ────────────────────────────────────────────────────────────────────

/** Get the full blockchain from a node */
export async function getChain(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/chain`);
}

/** Get blockchain length / metadata */
export async function getChainInfo(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/chain/info`);
}

// ─── Nodes ────────────────────────────────────────────────────────────────────

/** Get the peer nodes known to this node */
export async function getPeers(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/nodes`);
}

/** Register a new peer with a node */
export async function registerPeer(nodeUrl, peerUrl) {
  return fetchWithTimeout(`${nodeUrl}/nodes/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ nodes: [peerUrl] }),
  });
}

// ─── Mining & Transactions ────────────────────────────────────────────────────

/** Mine a new block on a node */
export async function mineBlock(nodeUrl, data = 'Mined via BlockView') {
  return fetchWithTimeout(`${nodeUrl}/mine`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ data }),
  });
}

/** Get pending transactions in the mempool */
export async function getPendingTransactions(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/transactions/pending`);
}

/** Submit a new transaction to a node */
export async function submitTransaction(nodeUrl, transaction) {
  return fetchWithTimeout(`${nodeUrl}/transactions/new`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(transaction),
  });
}

// ─── Consensus ────────────────────────────────────────────────────────────────

/** Trigger consensus / chain resolution on a node */
export async function resolveConflicts(nodeUrl) {
  return fetchWithTimeout(`${nodeUrl}/nodes/resolve`);
}
