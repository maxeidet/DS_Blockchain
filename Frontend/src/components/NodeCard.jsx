import React from 'react';

/**
 * NodeCard – Shows live status summary for a single GO_Blockchain node.
 *
 * Props:
 *   nodeUrl   {string}   – e.g. "http://localhost:8080"
 *   data      {object}   – { blocks, status, peers, mempool, online, loading, error, lastUpdated }
 *   onRemove  {function} – callback to remove this node
 *   selected  {boolean}
 *   onClick   {function}
 *
 * GO /status fields used:
 *   height, mempool_size, is_valid, difficulty, wallet_address, faucet_address
 */
export default function NodeCard({ nodeUrl, data, onRemove, selected, onClick }) {
  const { blocks, status, peers, mempool, online, loading, error, lastUpdated } = data;

  const blockCount   = status?.height != null ? status.height + 1 : (Array.isArray(blocks) ? blocks.length : '—');
  const peerCount    = Array.isArray(peers) ? peers.length : (status?.peer_count ?? '—');
  const pendingCount = status?.mempool_size ?? (Array.isArray(mempool) ? mempool.length : '—');

  return (
    <div
      className="card fade-in"
      onClick={onClick}
      style={{
        cursor: 'pointer',
        borderColor: selected ? 'var(--accent-primary)' : undefined,
        boxShadow: selected ? '0 0 20px var(--glow)' : undefined,
      }}
    >
      {/* Header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <span className={`dot ${online ? 'dot-green pulse' : 'dot-red'}`} />
          <code className="mono" style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>{nodeUrl}</code>
        </div>
        <button
          className="btn btn-ghost"
          style={{ padding: '4px 10px', fontSize: '0.75rem' }}
          onClick={(e) => { e.stopPropagation(); onRemove(nodeUrl); }}
        >
          ✕
        </button>
      </div>

      {/* Status badge */}
      {loading && !status && (
        <div style={{ color: 'var(--text-muted)', fontSize: '0.8rem', marginBottom: 10 }}>Connecting…</div>
      )}
      {error && (
        <div className="badge badge-red" style={{ marginBottom: 10 }}>{error}</div>
      )}

      {/* Extra info when online */}
      {status && (
        <div style={{ marginBottom: 10, display: 'flex', gap: 6, flexWrap: 'wrap' }}>
          <span className="badge badge-blue">diff {status.difficulty}</span>
          {status.is_valid === true  && <span className="badge badge-green">✓ valid</span>}
          {status.is_valid === false && <span className="badge badge-red">✗ invalid</span>}
        </div>
      )}

      {/* Stats grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
        <Stat label="Blocks"   value={blockCount}   color="var(--accent-primary)" />
        <Stat label="Peers"    value={peerCount}    color="var(--accent-secondary)" />
        <Stat label="Pending"  value={pendingCount} color="var(--accent-warning)" />
      </div>

      {/* Last updated */}
      {lastUpdated && (
        <div style={{ marginTop: 12, fontSize: '0.72rem', color: 'var(--text-muted)' }}>
          Updated {lastUpdated.toLocaleTimeString()}
        </div>
      )}
    </div>
  );
}

function Stat({ label, value, color }) {
  return (
    <div style={{ background: 'var(--bg-secondary)', borderRadius: 8, padding: '10px 12px', textAlign: 'center' }}>
      <div style={{ fontSize: '1.3rem', fontWeight: 700, color, fontFamily: 'var(--font-mono)' }}>{value}</div>
      <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', marginTop: 2 }}>{label}</div>
    </div>
  );
}
