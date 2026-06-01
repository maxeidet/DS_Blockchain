import React, { useState } from 'react';

/**
 * BlockChainView – Visualises GO_Blockchain's GET /blocks as a horizontal chain.
 *
 * Each block is rendered as a physical "card block" with a connector arrow
 * to the next. The genesis block is leftmost; latest block is rightmost.
 * Clicking a block opens a detail panel.
 *
 * Props:
 *   blocks {Block[]|null} – raw array from GET /blocks (oldest → newest)
 */
export default function BlockChainView({ blocks }) {
  const [selected, setSelected] = useState(null);

  const list = Array.isArray(blocks) ? blocks : [];

  if (!list.length) {
    return (
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        height: 200, color: 'var(--text-muted)', fontSize: '0.9rem',
        flexDirection: 'column', gap: 10,
      }}>
        <span style={{ fontSize: '2rem' }}></span>
        No block data yet. Select an online node.
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 0, height: '100%' }}>

      {/* ── Horizontal chain scroll area ── */}
      <div style={{
        overflowX: 'auto',
        overflowY: 'hidden',
        padding: '28px 24px 20px',
        display: 'flex',
        alignItems: 'center',
        gap: 0,
        flexShrink: 0,
        // Subtle dot-grid background for the "chain area"
        background: 'radial-gradient(circle, #e2e8f0 1px, transparent 1px) center / 22px 22px',
        borderBottom: '1px solid var(--border)',
        minHeight: 180,
      }}>
        {list.map((block, i) => {
          const isLatest = i === list.length - 1;
          const isGenesis = i === 0;
          const isSelected = selected?.idx === i;
          return (
            <React.Fragment key={block.Index ?? block.index ?? i}>
              <BlockCard
                block={block}
                isLatest={isLatest}
                isGenesis={isGenesis}
                isSelected={isSelected}
                onClick={() => setSelected(isSelected ? null : { block, idx: i })}
              />
              {/* Arrow connector between blocks */}
              {i < list.length - 1 && <ChainConnector />}
            </React.Fragment>
          );
        })}
      </div>

      {/* ── Detail panel ── */}
      {selected ? (
        <BlockDetailPanel
          block={selected.block}
          isLatest={selected.idx === list.length - 1}
          onClose={() => setSelected(null)}
        />
      ) : (
        <div style={{
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          flex: 1, color: 'var(--text-muted)', fontSize: '0.85rem', gap: 6,
        }}>
          Click a block to inspect it
        </div>
      )}
    </div>
  );
}

// ─── Block Card ──────────────────────────────────────────────────────────────

function BlockCard({ block, isLatest, isGenesis, isSelected, onClick }) {
  const index = block.Index ?? block.index ?? '?';
  const hash = block.Hash ?? block.hash ?? '';
  const txCount = (block.Transactions ?? block.transactions ?? []).length;
  const ts = block.Timestamp ?? block.timestamp;
  const nonce = block.Nonce ?? block.nonce;

  const shortHash = hash ? `${hash.slice(0, 6)}…${hash.slice(-4)}` : '—';
  const timeStr = ts ? new Date(ts * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';

  // Color accent per block category
  const accentColor = isGenesis ? 'var(--accent-secondary)'
    : isLatest ? 'var(--accent-success)'
      : 'var(--accent-primary)';

  return (
    <div
      onClick={onClick}
      style={{
        width: 132,
        flexShrink: 0,
        background: '#fff',
        border: `2px solid ${isSelected ? accentColor : 'var(--border)'}`,
        borderRadius: 14,
        padding: '14px 12px',
        cursor: 'pointer',
        boxShadow: isSelected
          ? `0 0 0 3px ${accentColor}22, var(--shadow-md)`
          : 'var(--shadow-sm)',
        transform: isSelected ? 'translateY(-4px) scale(1.04)' : 'scale(1)',
        transition: 'all 0.2s cubic-bezier(0.34,1.56,0.64,1)',
        position: 'relative',
        userSelect: 'none',
      }}
    >
      {/* Top stripe accent */}
      <div style={{
        position: 'absolute', top: 0, left: 12, right: 12, height: 3,
        background: accentColor,
        borderRadius: '0 0 4px 4px',
      }} />

      {/* Block index */}
      <div style={{
        fontSize: '1.25rem', fontWeight: 700,
        color: accentColor,
        fontFamily: 'var(--font-mono)',
        lineHeight: 1,
        marginBottom: 6,
        marginTop: 4,
      }}>
        #{index}
      </div>

      {/* Label */}
      <div style={{ fontSize: '0.65rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.07em', color: 'var(--text-muted)', marginBottom: 10 }}>
        {isGenesis ? 'Genesis' : isLatest ? 'Latest' : 'Block'}
      </div>

      {/* Hash pill */}
      <div style={{
        background: 'var(--bg-muted)',
        borderRadius: 6,
        padding: '3px 6px',
        fontFamily: 'var(--font-mono)',
        fontSize: '0.62rem',
        color: 'var(--text-secondary)',
        marginBottom: 8,
        letterSpacing: '-0.02em',
      }}>
        {shortHash}
      </div>

      {/* Stats row */}
      <div style={{ display: 'flex', gap: 4, flexWrap: 'wrap' }}>
        <MiniTag value={`${txCount} tx`} />
        {nonce != null && <MiniTag value={`n:${nonce}`} />}
      </div>

      {/* Time */}
      {timeStr && (
        <div style={{ marginTop: 8, fontSize: '0.62rem', color: 'var(--text-muted)' }}>
          {timeStr}
        </div>
      )}
    </div>
  );
}

function MiniTag({ value }) {
  return (
    <span style={{
      background: 'var(--bg-muted)',
      border: '1px solid var(--border)',
      borderRadius: 4,
      padding: '1px 5px',
      fontSize: '0.6rem',
      color: 'var(--text-secondary)',
      fontFamily: 'var(--font-mono)',
    }}>
      {value}
    </span>
  );
}

// ─── Chain connector arrow ────────────────────────────────────────────────────

function ChainConnector() {
  return (
    <div style={{
      display: 'flex', alignItems: 'center',
      flexShrink: 0, padding: '0 2px',
      position: 'relative',
    }}>
      {/* Dashed chain link */}
      <div style={{
        width: 36, height: 2,
        background: 'repeating-linear-gradient(90deg, var(--border) 0 6px, transparent 6px 10px)',
        position: 'relative',
      }} />
      {/* Arrowhead */}
      <div style={{
        width: 0, height: 0,
        borderTop: '5px solid transparent',
        borderBottom: '5px solid transparent',
        borderLeft: '7px solid var(--border)',
        marginLeft: -1,
      }} />
    </div>
  );
}

// ─── Block Detail Panel ───────────────────────────────────────────────────────

function BlockDetailPanel({ block, isLatest, onClose }) {
  const index = block.Index ?? block.index ?? '?';
  const hash = block.Hash ?? block.hash;
  const prevHash = block.PrevHash ?? block.prev_hash ?? block.previous_hash;
  const nonce = block.Nonce ?? block.nonce;
  const difficulty = block.Difficulty ?? block.difficulty;
  const ts = block.Timestamp ?? block.timestamp;
  const miner = block.MinerAddress ?? block.miner_address;
  const reward = block.MiningReward ?? block.mining_reward;
  const txs = block.Transactions ?? block.transactions ?? [];

  const timeStr = ts ? new Date(ts * 1000).toLocaleString() : null;

  return (
    <div className="slide-in" style={{
      flex: 1,
      overflowY: 'auto',
      padding: '24px',
      background: 'var(--bg-primary)',
    }}>
      {/* Panel header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{
            width: 40, height: 40,
            background: isLatest ? 'var(--accent-success)' : 'var(--accent-primary)',
            borderRadius: 10,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            color: '#fff', fontWeight: 700, fontSize: '1rem',
            fontFamily: 'var(--font-mono)',
          }}>
            #{index}
          </div>
          <div>
            <div style={{ fontWeight: 700, fontSize: '1.05rem' }}>Block #{index}</div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              {isLatest && <span style={{ color: 'var(--accent-success)', marginRight: 6 }}>● Latest</span>}
              {timeStr}
            </div>
          </div>
        </div>
        <button className="btn btn-ghost" onClick={onClose} style={{ padding: '6px 12px' }}>Close</button>
      </div>

      {/* Main grid */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14, marginBottom: 14 }}>
        <DetailCard label="Nonce" value={nonce ?? '—'} mono />
        <DetailCard label="Difficulty" value={difficulty ?? '—'} />
        {miner && <DetailCard label="Miner Address" value={miner} mono span2 />}
        {reward != null && <DetailCard label="Mining Reward" value={`${reward} tokens`} />}
        {timeStr && <DetailCard label="Timestamp" value={timeStr} />}
      </div>

      {/* Hashes */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 14 }}>
        <HashRow label="Hash" value={hash} color="var(--accent-primary)" />
        <HashRow label="Previous Hash" value={prevHash} color="var(--text-muted)" />
      </div>

      {/* Transactions */}
      <div>
        <div style={{
          display: 'flex', alignItems: 'center', gap: 8, marginBottom: 12,
          fontWeight: 600, fontSize: '0.85rem',
        }}>
          Transactions
          <span className="badge badge-blue">{txs.length}</span>
        </div>

        {txs.length === 0 ? (
          <div style={{
            background: 'var(--bg-muted)', borderRadius: 10, padding: '20px',
            color: 'var(--text-muted)', fontSize: '0.82rem', textAlign: 'center',
          }}>
            No transactions in this block
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {txs.map((tx, i) => <TxRow key={tx.ID ?? tx.id ?? i} tx={tx} index={i} />)}
          </div>
        )}
      </div>
    </div>
  );
}

// ─── Sub-components ──────────────────────────────────────────────────────────

function DetailCard({ label, value, mono, span2 }) {
  return (
    <div style={{
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: 10,
      padding: '12px 14px',
      gridColumn: span2 ? '1 / -1' : undefined,
      boxShadow: 'var(--shadow-sm)',
    }}>
      <div style={{ fontSize: '0.67rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.07em', color: 'var(--text-muted)', marginBottom: 4 }}>
        {label}
      </div>
      <div style={{
        fontFamily: mono ? 'var(--font-mono)' : undefined,
        fontSize: mono ? '0.78rem' : '0.9rem',
        fontWeight: 600,
        color: 'var(--text-primary)',
        wordBreak: 'break-all',
      }}>
        {String(value)}
      </div>
    </div>
  );
}

function HashRow({ label, value, color }) {
  if (!value) return null;
  return (
    <div style={{
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: 10,
      padding: '12px 14px',
      boxShadow: 'var(--shadow-sm)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 6 }}>
        <span style={{ width: 8, height: 8, borderRadius: '50%', background: color, display: 'inline-block' }} />
        <span style={{ fontSize: '0.67rem', fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.07em', color: 'var(--text-muted)' }}>
          {label}
        </span>
      </div>
      <code style={{
        fontFamily: 'var(--font-mono)',
        fontSize: '0.75rem',
        color: 'var(--text-secondary)',
        wordBreak: 'break-all',
        display: 'block',
        lineHeight: 1.6,
      }}>
        {value}
      </code>
    </div>
  );
}

function TxRow({ tx, index }) {
  const [open, setOpen] = useState(false);

  const from = tx.From ?? tx.from ?? tx.sender ?? '?';
  const to = tx.To ?? tx.to ?? tx.recipient ?? '?';
  const amount = tx.Amount ?? tx.amount ?? '?';
  const txType = tx.Type ?? tx.type ?? 'transfer';
  const nonce = tx.Nonce ?? tx.nonce;

  const short = (addr) => addr.length > 12 ? `${addr.slice(0, 8)}…${addr.slice(-4)}` : addr;

  return (
    <div style={{
      background: '#fff',
      border: '1px solid var(--border)',
      borderRadius: 10,
      overflow: 'hidden',
      boxShadow: 'var(--shadow-sm)',
    }}>
      {/* Summary row */}
      <div
        onClick={() => setOpen(o => !o)}
        style={{
          display: 'flex', alignItems: 'center', gap: 12,
          padding: '10px 14px', cursor: 'pointer',
          background: open ? 'var(--bg-muted)' : '#fff',
          transition: 'background 0.15s',
        }}
      >
        <span style={{
          width: 24, height: 24, borderRadius: 6,
          background: 'var(--accent-primary)',
          color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontSize: '0.65rem', fontWeight: 700, flexShrink: 0,
        }}>
          {index + 1}
        </span>
        <div style={{ flex: 1, overflow: 'hidden' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.8rem' }}>
            <code className="mono" style={{ fontSize: '0.72rem', color: 'var(--text-secondary)' }}>{short(from)}</code>
            <span style={{ color: 'var(--text-muted)' }}>→</span>
            <code className="mono" style={{ fontSize: '0.72rem', color: 'var(--text-secondary)' }}>{short(to)}</code>
          </div>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
          <span style={{ fontWeight: 700, fontSize: '0.88rem', color: 'var(--accent-primary)' }}>
            {amount} <span style={{ fontSize: '0.65rem', fontWeight: 500, color: 'var(--text-muted)' }}>tokens</span>
          </span>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>{open ? '▲' : '▼'}</span>
        </div>
      </div>

      {/* Expanded detail */}
      {open && (
        <div style={{
          borderTop: '1px solid var(--border)',
          padding: '12px 14px',
          display: 'flex', flexDirection: 'column', gap: 6,
          fontSize: '0.78rem',
        }}>
          <TxField label="From" value={from} mono />
          <TxField label="To" value={to} mono />
          <TxField label="Amount" value={`${amount} tokens`} />
          <TxField label="Type" value={txType} />
          {nonce != null && <TxField label="Nonce" value={nonce} />}
          {(tx.Signature ?? tx.signature) && (
            <TxField label="Signature" value={tx.Signature ?? tx.signature} mono />
          )}
        </div>
      )}
    </div>
  );
}

function TxField({ label, value, mono }) {
  return (
    <div style={{ display: 'flex', gap: 8 }}>
      <span style={{ color: 'var(--text-muted)', minWidth: 70, flexShrink: 0 }}>{label}</span>
      <span style={{
        fontFamily: mono ? 'var(--font-mono)' : undefined,
        color: 'var(--text-primary)',
        wordBreak: 'break-all',
        fontSize: mono ? '0.72rem' : undefined,
      }}>
        {String(value)}
      </span>
    </div>
  );
}
