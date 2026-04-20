import React, { useState } from 'react';
import { submitTransaction, mineBlock, resolveConflicts } from '../api/blockchain';

/**
 * NodeActions – control panel for a selected node.
 *
 * Props:
 *   nodeUrl {string}
 *   onAction {function} – called after any action to trigger a data refresh
 */
export default function NodeActions({ nodeUrl, onAction }) {
  const [sender, setSender]     = useState('');
  const [recipient, setRecipient] = useState('');
  const [amount, setAmount]     = useState('');
  const [status, setStatus]     = useState(null);
  const [busy, setBusy]         = useState(false);

  async function run(fn, label) {
    setBusy(true);
    setStatus({ type: 'info', msg: `${label}…` });
    try {
      const res = await fn();
      setStatus({ type: 'success', msg: JSON.stringify(res, null, 2) });
      onAction?.();
    } catch (err) {
      setStatus({ type: 'error', msg: err.message });
    } finally {
      setBusy(false);
    }
  }

  function handleSubmitTx(e) {
    e.preventDefault();
    run(() => submitTransaction(nodeUrl, { sender, recipient, amount: Number(amount) }), 'Submitting transaction');
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* Transaction Form */}
      <div className="card">
        <h3 style={{ marginBottom: 14, fontSize: '0.95rem', color: 'var(--text-primary)' }}>New Transaction</h3>
        <form onSubmit={handleSubmitTx} style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          <input placeholder="Sender" value={sender}    onChange={e => setSender(e.target.value)}    required />
          <input placeholder="Recipient" value={recipient} onChange={e => setRecipient(e.target.value)} required />
          <input placeholder="Amount" type="number" value={amount} onChange={e => setAmount(e.target.value)} required />
          <button className="btn btn-primary" type="submit" disabled={busy}>Submit Transaction</button>
        </form>
      </div>

      {/* Mine */}
      <div className="card">
        <h3 style={{ marginBottom: 12, fontSize: '0.95rem' }}>Mining</h3>
        <button
          className="btn btn-primary"
          style={{ width: '100%' }}
          disabled={busy}
          onClick={() => run(() => mineBlock(nodeUrl), 'Mining block')}
        >
          ⛏ Mine Block
        </button>
      </div>

      {/* Consensus */}
      <div className="card">
        <h3 style={{ marginBottom: 12, fontSize: '0.95rem' }}>Consensus</h3>
        <button
          className="btn btn-ghost"
          style={{ width: '100%' }}
          disabled={busy}
          onClick={() => run(() => resolveConflicts(nodeUrl), 'Resolving conflicts')}
        >
          ⚖ Resolve Conflicts
        </button>
      </div>

      {/* Status output */}
      {status && (
        <div
          className="card fade-in"
          style={{
            borderColor: status.type === 'success' ? 'var(--accent-success)'
                       : status.type === 'error'   ? 'var(--accent-danger)'
                       : 'var(--accent-primary)',
          }}
        >
          <pre className="mono" style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {status.msg}
          </pre>
        </div>
      )}
    </div>
  );
}
