import React from 'react';

/**
 * BlockList – Renders the blocks in a chain.
 *
 * Props:
 *   chain {object|null} – raw response from GET /chain
 */
export default function BlockList({ chain }) {
  const blocks = chain?.chain ?? (Array.isArray(chain) ? chain : []);

  if (!blocks.length) {
    return (
      <div style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '40px 0' }}>
        No chain data yet. Select an online node.
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      {[...blocks].reverse().map((block, i) => (
        <BlockItem key={block.index ?? i} block={block} isLatest={i === 0} />
      ))}
    </div>
  );
}

function BlockItem({ block, isLatest }) {
  const [expanded, setExpanded] = React.useState(false);

  return (
    <div
      className="card fade-in"
      style={{ padding: '14px 18px', borderColor: isLatest ? 'var(--accent-success)' : undefined }}
    >
      <div
        style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', cursor: 'pointer' }}
        onClick={() => setExpanded(v => !v)}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          {isLatest && <span className="badge badge-green">Latest</span>}
          <span style={{ fontWeight: 600 }}>Block #{block.index ?? '?'}</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <span style={{ fontSize: '0.78rem', color: 'var(--text-muted)' }}>
            {block.timestamp ? new Date(block.timestamp * 1000).toLocaleString() : ''}
          </span>
          <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>{expanded ? '▲' : '▼'}</span>
        </div>
      </div>

      {expanded && (
        <div style={{ marginTop: 14, display: 'flex', flexDirection: 'column', gap: 8 }}>
          <Field label="Hash"          value={block.hash ?? block.current_hash} mono />
          <Field label="Previous Hash" value={block.previous_hash}              mono />
          <Field label="Nonce"         value={block.nonce} />
          <Field label="Transactions"  value={JSON.stringify(block.transactions, null, 2)} mono block />
        </div>
      )}
    </div>
  );
}

function Field({ label, value, mono, block }) {
  if (value === undefined || value === null) return null;
  return (
    <div>
      <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.05em', marginBottom: 2 }}>{label}</div>
      <pre
        className={mono ? 'mono' : ''}
        style={{
          fontSize: '0.78rem',
          color: 'var(--text-secondary)',
          wordBreak: 'break-all',
          whiteSpace: block ? 'pre-wrap' : 'nowrap',
          overflow: 'hidden',
          textOverflow: block ? undefined : 'ellipsis',
          background: 'var(--bg-secondary)',
          borderRadius: 6,
          padding: block ? '8px 10px' : '4px 8px',
        }}
      >
        {String(value)}
      </pre>
    </div>
  );
}
