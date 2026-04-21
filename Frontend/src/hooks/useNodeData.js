import { useState, useEffect, useCallback, useRef } from 'react';
import { getStatus, getBlocks, getMempool, getPeers } from '../api/blockchain';

const DEFAULT_INTERVAL = 5000; // ms between polls

/**
 * useNodeData
 *
 * Polls a single GO_Blockchain node for its current state.
 *
 * GO_Blockchain API:
 *   GET /status  → { height, difficulty, mining_reward, mempool_size, is_valid,
 *                    wallet_address, wallet_nonce, faucet_address }
 *   GET /blocks  → Block[]
 *   GET /mempool → Transaction[]
 *   GET /peers   → string[]
 *
 * @param {string} nodeUrl  – e.g. "http://localhost:8080"
 * @param {number} interval – polling interval in milliseconds
 */
export function useNodeData(nodeUrl, interval = DEFAULT_INTERVAL) {
  const [blocks, setBlocks]             = useState(null);   // Block[]
  const [status, setStatus]             = useState(null);   // GO /status object
  const [peers, setPeers]               = useState([]);     // string[]
  const [mempool, setMempool]           = useState([]);     // Transaction[]
  const [loading, setLoading]           = useState(false);
  const [error, setError]               = useState(null);
  const [lastUpdated, setLastUpdated]   = useState(null);
  const [online, setOnline]             = useState(false);

  const intervalRef = useRef(null);

  const fetchAll = useCallback(async () => {
    if (!nodeUrl) return;
    setLoading(true);
    try {
      // Fetch all endpoints in parallel; tolerate individual failures
      const [statusRes, blocksRes, mempoolRes, peersRes] = await Promise.allSettled([
        getStatus(nodeUrl),
        getBlocks(nodeUrl),
        getMempool(nodeUrl),
        getPeers(nodeUrl),
      ]);

      if (statusRes.status  === 'fulfilled') setStatus(statusRes.value);
      if (blocksRes.status  === 'fulfilled') setBlocks(blocksRes.value);
      if (mempoolRes.status === 'fulfilled') setMempool(Array.isArray(mempoolRes.value) ? mempoolRes.value : []);
      if (peersRes.status   === 'fulfilled') setPeers(Array.isArray(peersRes.value) ? peersRes.value : []);

      // Node is online if /status responded OK
      const isOnline = statusRes.status === 'fulfilled';
      setOnline(isOnline);
      setError(isOnline ? null : (statusRes.reason?.message ?? 'Unreachable'));
    } catch (err) {
      setOnline(false);
      setError(err.message);
    } finally {
      setLoading(false);
      setLastUpdated(new Date());
    }
  }, [nodeUrl]);

  // Initial fetch + polling
  useEffect(() => {
    fetchAll();
    intervalRef.current = setInterval(fetchAll, interval);
    return () => clearInterval(intervalRef.current);
  }, [fetchAll, interval]);

  return { blocks, status, peers, mempool, loading, error, lastUpdated, online, refetch: fetchAll };
}
