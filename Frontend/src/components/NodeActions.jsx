import React, { useState, useCallback } from 'react';
import { submitTransaction, mineBlock, faucetDrop } from '../api/blockchain';

/**
 * NodeActions – control panel for a selected GO_Blockchain node.
 *
 * Supported actions (matching GO_Blockchain's API):
 *   POST /transactions  { to, amount }         — node signs with its own wallet
 *   POST /faucet        { to, amount }          — faucet wallet sends tokens
 *   POST /mine          { miner_address }        — mine pending transactions
 *
 * Props:
 *   nodeUrl   {string}
 *   data      {object} – includes status.wallet_address, status.faucet_address
 *   onAction  {function} – called after any action to trigger a data refresh
 */

// ─── Per-button feedback states ──────────────────────────────────────────────
// 'idle' | 'loading' | 'success' | 'error'

const BTN_RESET_MS = 2000; // how long to show success/error before going back to idle

function useButtonState() {
  const [state, setState] = useState('idle');

  const run = useCallback(async (fn) => {
    setState('loading');
    try {
      const result = await fn();
      setState('success');
      setTimeout(() => setState('idle'), BTN_RESET_MS);
      return { ok: true, result };
    } catch (err) {
      setState('error');
      setTimeout(() => setState('idle'), BTN_RESET_MS);
      return { ok: false, error: err };
    }
  }, []);

  return [state, run];
}

// ─── ActionButton ─────────────────────────────────────────────────────────────
function ActionButton({ btnState, onClick, type = 'button', children, style = {} }) {
  const isLoading = btnState === 'loading';
  const isSuccess = btnState === 'success';
  const isError   = btnState === 'error';

  const bgColor = isSuccess ? 'var(--accent-success, #22c55e)'
                : isError   ? 'var(--accent-danger,  #ef4444)'
                : undefined;

  const label = isLoading ? <Spinner />
              : isSuccess ? 'Done'
              : isError   ? 'Failed'
              : children;

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={isLoading}
      style={{
        ...style,
        background: bgColor,
        borderColor: bgColor,
        color: (isSuccess || isError) ? '#fff' : undefined,
        transition: 'background 0.2s, border-color 0.2s, transform 0.1s',
        transform: isLoading ? 'scale(0.97)' : 'scale(1)',
      }}
      className="btn btn-primary"
    >
      {label}
    </button>
  );
}

function Spinner() {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
      <span style={{
        width: 12, height: 12,
        border: '2px solid rgba(255,255,255,0.4)',
        borderTopColor: '#fff',
        borderRadius: '50%',
        display: 'inline-block',
        animation: 'spin 0.7s linear infinite',
      }} />
      Processing…
    </span>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────
export default function NodeActions({ nodeUrl, data, onAction }) {
  const { status } = data ?? {};

  // Transaction state
  const [txTo, setTxTo]         = useState('');
  const [txAmount, setTxAmount] = useState('');
  const [txBtn, runTx]          = useButtonState();

  // Faucet state
  const [faucetTo, setFaucetTo]         = useState('');
  const [faucetAmount, setFaucetAmount] = useState('');
  const [faucetBtn, runFaucet]          = useButtonState();

  // Mine state
  const [minerAddr, setMinerAddr] = useState(status?.wallet_address ?? '');
  const [mineBtn, runMine]        = useButtonState();

  // Result panel (shows last response)
  const [lastResult, setLastResult] = useState(null);

  // Keep miner field in sync when status loads
  React.useEffect(() => {
    if (status?.wallet_address) {
      setMinerAddr(current => current || status.wallet_address);
    }
  }, [status?.wallet_address]);

  async function handleSubmitTx(e) {
    e.preventDefault();
    const { ok, result, error } = await runTx(
      () => submitTransaction(nodeUrl, { to: txTo.trim(), amount: Number(txAmount) }),
    );
    setLastResult({ ok, msg: ok ? JSON.stringify(result, null, 2) : error.message });
    if (ok) onAction?.();
  }

  async function handleFaucet(e) {
    e.preventDefault();
    const { ok, result, error } = await runFaucet(
      () => faucetDrop(nodeUrl, { to: faucetTo.trim(), amount: Number(faucetAmount) }),
    );
    setLastResult({ ok, msg: ok ? JSON.stringify(result, null, 2) : error.message });
    if (ok) onAction?.();
  }

  async function handleMine() {
    const { ok, result, error } = await runMine(
      () => mineBlock(nodeUrl, minerAddr.trim()),
    );
    setLastResult({ ok, msg: ok ? JSON.stringify(result, null, 2) : error.message });
    if (ok) onAction?.();
  }

  const anyBusy = txBtn === 'loading' || faucetBtn === 'loading' || mineBtn === 'loading';

  return (
    <>
      {/* keyframe for the spinner — injected once */}
      <style>{`@keyframes spin { to { transform: rotate(360deg); } }`}</style>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

        {/* Node info strip */}
        {status && (
          <div className="card" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 6 }}>
            <InfoRow label="Node wallet"  value={status.wallet_address} />
            <InfoRow label="Faucet"       value={status.faucet_address} />
            <InfoRow label="Chain height" value={status.height} />
            <InfoRow label="Mempool"      value={`${status.mempool_size} pending`} />
          </div>
        )}

        {/* Transaction Form */}
        <div className="card">
          <h3 style={{ marginBottom: 14, fontSize: '0.95rem', color: 'var(--text-primary)' }}>
            New Transaction
            <span style={{ fontSize: '0.72rem', fontWeight: 400, color: 'var(--text-muted)', marginLeft: 8 }}>
              (signed by node wallet)
            </span>
          </h3>
          <form onSubmit={handleSubmitTx} style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <input
              placeholder="To (recipient address)"
              value={txTo}
              onChange={e => setTxTo(e.target.value)}
              required
              disabled={anyBusy}
            />
            <input
              placeholder="Amount"
              type="number"
              min="1"
              value={txAmount}
              onChange={e => setTxAmount(e.target.value)}
              required
              disabled={anyBusy}
            />
            <ActionButton btnState={txBtn} type="submit">Submit Transaction</ActionButton>
          </form>
        </div>

        {/* Faucet Form */}
        <div className="card">
          <h3 style={{ marginBottom: 14, fontSize: '0.95rem', color: 'var(--text-primary)' }}>
            Faucet
            <span style={{ fontSize: '0.72rem', fontWeight: 400, color: 'var(--text-muted)', marginLeft: 8 }}>
              (request tokens from faucet wallet)
            </span>
          </h3>
          <form onSubmit={handleFaucet} style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            <input
              placeholder="To (your address)"
              value={faucetTo}
              onChange={e => setFaucetTo(e.target.value)}
              required
              disabled={anyBusy}
            />
            <input
              placeholder="Amount"
              type="number"
              min="1"
              value={faucetAmount}
              onChange={e => setFaucetAmount(e.target.value)}
              required
              disabled={anyBusy}
            />
            <ActionButton btnState={faucetBtn} type="submit">Request Tokens</ActionButton>
          </form>
        </div>

        {/* Mine */}
        <div className="card">
          <h3 style={{ marginBottom: 12, fontSize: '0.95rem' }}>Mining</h3>
          <div style={{ display: 'flex', gap: 8 }}>
            <input
              placeholder="Miner address (defaults to node wallet)"
              value={minerAddr}
              onChange={e => setMinerAddr(e.target.value)}
              style={{ flex: 1 }}
              disabled={anyBusy}
            />
            <ActionButton
              btnState={mineBtn}
              onClick={handleMine}
              style={{ flexShrink: 0 }}
            >
              Mine Block
            </ActionButton>
          </div>
        </div>

        {/* Last result panel */}
        {lastResult && (
          <div
            className="card fade-in"
            style={{
              borderColor: lastResult.ok ? 'var(--accent-success)' : 'var(--accent-danger)',
            }}
          >
            <pre className="mono" style={{
              fontSize: '0.78rem',
              color: 'var(--text-secondary)',
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-all',
            }}>
              {lastResult.msg}
            </pre>
          </div>
        )}
      </div>
    </>
  );
}

function InfoRow({ label, value }) {
  if (value == null || value === '') return null;
  return (
    <div style={{ display: 'flex', gap: 8, alignItems: 'baseline', fontSize: '0.8rem' }}>
      <span style={{ color: 'var(--text-muted)', minWidth: 90 }}>{label}</span>
      <code className="mono" style={{ color: 'var(--text-secondary)', wordBreak: 'break-all' }}>{String(value)}</code>
    </div>
  );
}
