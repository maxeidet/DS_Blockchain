import { useState, useEffect, useCallback, useRef } from 'react';
import { getChain, getPeers, getPendingTransactions } from '../api/blockchain';

const DEFAULT_INTERVAL = 5000; // ms between polls

/**
 * useNodeData
 *
 * Polls a single blockchain node for its current state.
 *
 * @param {string} nodeUrl  – e.g. "http://localhost:5001"
 * @param {number} interval – polling interval in milliseconds
 */
export function useNodeData(nodeUrl, interval = DEFAULT_INTERVAL) {
  const [chain, setChain]               = useState(null);
  const [peers, setPeers]               = useState([]);
  const [pending, setPending]           = useState([]);
  const [loading, setLoading]           = useState(false);
  const [error, setError]               = useState(null);
  const [lastUpdated, setLastUpdated]   = useState(null);
  const [online, setOnline]             = useState(false);

  const intervalRef = useRef(null);

  const fetchAll = useCallback(async () => {
    if (!nodeUrl) return;
    setLoading(true);
    try {
      // Fetch in parallel
      const [chainData, peersData, pendingData] = await Promise.allSettled([
        getChain(nodeUrl),
        getPeers(nodeUrl),
        getPendingTransactions(nodeUrl),
      ]);

      if (chainData.status === 'fulfilled')   setChain(chainData.value);
      if (peersData.status === 'fulfilled')   setPeers(peersData.value?.nodes ?? peersData.value ?? []);
      if (pendingData.status === 'fulfilled') setPending(pendingData.value?.transactions ?? pendingData.value ?? []);

      // If at least the chain responded, consider node online
      setOnline(chainData.status === 'fulfilled');
      setError(chainData.status === 'rejected' ? chainData.reason?.message : null);
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

  return { chain, peers, pending, loading, error, lastUpdated, online, refetch: fetchAll };
}
