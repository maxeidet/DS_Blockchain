import { useState, useCallback } from 'react';

const STORAGE_KEY = 'blockchain_nodes';

function loadNodes() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? JSON.parse(stored) : ['http://localhost:8080'];
  } catch {
    return ['http://localhost:8080'];
  }
}

/**
 * useNodeRegistry
 *
 * Manages the frontend's known set of blockchain nodes.
 * Persists to localStorage so nodes survive page refreshes.
 */
export function useNodeRegistry() {
  const [nodes, setNodes] = useState(loadNodes);

  const persist = useCallback((updated) => {
    setNodes(updated);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(updated));
  }, []);

  const addNode = useCallback((url) => {
    const trimmed = url.trim().replace(/\/$/, ''); // strip trailing slash
    if (!trimmed || nodes.includes(trimmed)) return;
    persist([...nodes, trimmed]);
  }, [nodes, persist]);

  const removeNode = useCallback((url) => {
    persist(nodes.filter(n => n !== url));
  }, [nodes, persist]);

  return { nodes, addNode, removeNode };
}
