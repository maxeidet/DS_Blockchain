import React, { useMemo, useState } from 'react';

function getField(tx, ...keys) {
  for (const key of keys) {
    if (tx?.[key] != null && tx[key] !== '') return tx[key];
  }
  return '';
}

function shortValue(value, head = 10, tail = 6) {
  const text = String(value || '');
  if (!text) return '-';
  return text.length > head + tail + 3 ? `${text.slice(0, head)}...${text.slice(-tail)}` : text;
}

function formatTimestamp(value) {
  if (!value) return 'Unknown time';
  const numeric = Number(value);
  const date = new Date(numeric > 10_000_000_000 ? numeric : numeric * 1000);
  if (Number.isNaN(date.getTime())) return 'Unknown time';
  return date.toLocaleString();
}

function txTypeClass(type) {
  switch (String(type || '').toLowerCase()) {
    case 'faucet':
      return 'badge-purple';
    case 'coinbase':
      return 'badge-green';
    case 'transfer':
      return 'badge-blue';
    default:
      return 'badge-amber';
  }
}

function CopyValue({ label, value, color = 'var(--accent-primary)' }) {
  const [copied, setCopied] = useState(false);
  const text = String(value || '');

  async function handleCopy() {
    if (!text || !navigator?.clipboard) return;
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1400);
  }

  return (
    <button
      type="button"
      onClick={handleCopy}
      title={`Copy ${label}`}
      style={{
        minWidth: 0,
        display: 'flex',
        alignItems: 'center',
        gap: 8,
        background: '#fff',
        border: '1px solid var(--border)',
        borderLeft: `4px solid ${color}`,
        borderRadius: 8,
        padding: '9px 10px',
        textAlign: 'left',
      }}
    >
      <span style={{ minWidth: 68, color: 'var(--text-muted)', fontSize: '0.68rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.05em' }}>
        {label}
      </span>
      <code className="mono truncate" style={{ flex: 1, color: 'var(--text-secondary)', fontSize: '0.76rem' }}>
        {shortValue(text)}
      </code>
      <span style={{ color: copied ? 'var(--accent-success)' : 'var(--text-muted)', fontSize: '0.72rem', flexShrink: 0 }}>
        {copied ? 'Copied' : 'Copy'}
      </span>
    </button>
  );
}

function Detail({ label, value, mono = false }) {
  if (value == null || value === '') return null;
  return (
    <div style={{
      background: 'var(--bg-muted)',
      border: '1px solid var(--border)',
      borderRadius: 8,
      padding: '10px 12px',
      minWidth: 0,
    }}>
      <div style={{ color: 'var(--text-muted)', fontSize: '0.67rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', marginBottom: 4 }}>
        {label}
      </div>
      <div
        className={mono ? 'mono' : undefined}
        style={{
          color: 'var(--text-primary)',
          fontSize: mono ? '0.72rem' : '0.86rem',
          fontWeight: mono ? 500 : 650,
          wordBreak: 'break-all',
        }}
      >
        {String(value)}
      </div>
    </div>
  );
}

function TransactionCard({ tx, index }) {
  const id = getField(tx, 'id', 'ID');
  const from = getField(tx, 'from', 'From', 'sender');
  const to = getField(tx, 'to', 'To', 'recipient');
  const amount = getField(tx, 'amount', 'Amount');
  const timestamp = getField(tx, 'timestamp', 'Timestamp');
  const nonce = getField(tx, 'nonce', 'Nonce');
  const type = getField(tx, 'type', 'Type') || 'transfer';
  const signature = getField(tx, 'signature', 'Signature');
  const publicKey = getField(tx, 'public_key', 'PublicKey', 'publicKey');

  return (
    <article className="card fade-in" style={{ padding: 0, overflow: 'hidden', borderRadius: 10 }}>
      <div style={{
        padding: '14px 16px',
        display: 'grid',
        gridTemplateColumns: 'auto 1fr auto',
        gap: 14,
        alignItems: 'center',
        borderBottom: '1px solid var(--border)',
        background: '#fff',
      }}>
        <div style={{
          width: 34,
          height: 34,
          borderRadius: 8,
          background: 'var(--accent-primary)',
          color: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontWeight: 800,
          fontSize: '0.82rem',
          fontFamily: 'var(--font-mono)',
        }}>
          {index + 1}
        </div>

        <div style={{ minWidth: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap', marginBottom: 4 }}>
            <span className={`badge ${txTypeClass(type)}`}>{String(type)}</span>
            <span style={{ color: 'var(--text-muted)', fontSize: '0.78rem' }}>{formatTimestamp(timestamp)}</span>
          </div>
          <code className="mono" style={{ color: 'var(--text-secondary)', fontSize: '0.73rem', wordBreak: 'break-all' }}>
            {id || 'Transaction ID pending'}
          </code>
        </div>

        <div style={{
          background: '#fff7ed',
          border: '1px solid #fed7aa',
          color: '#c2410c',
          borderRadius: 999,
          padding: '7px 12px',
          fontWeight: 800,
          fontSize: '0.95rem',
          whiteSpace: 'nowrap',
        }}>
          {amount || 0} <span style={{ fontSize: '0.68rem', fontWeight: 700 }}>tokens</span>
        </div>
      </div>

      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1fr) minmax(0, 1fr)', gap: 10 }}>
          <CopyValue label="Sender" value={from} color="var(--accent-secondary)" />
          <CopyValue label="Recipient" value={to} color="var(--accent-success)" />
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: 10 }}>
          <Detail label="Nonce" value={nonce || '0'} />
          <Detail label="Signature" value={signature || 'Unsigned'} mono />
          <Detail label="Public Key" value={publicKey || 'Not included'} mono />
        </div>
      </div>
    </article>
  );
}

export default function MempoolView({ mempool, nodeUrl, onAction, onGoToActions }) {
  const [query, setQuery] = useState('');
  const transactions = useMemo(() => (Array.isArray(mempool) ? mempool : []), [mempool]);
  const normalizedQuery = query.trim().toLowerCase();

  const filtered = useMemo(() => {
    if (!normalizedQuery) return transactions;
    return transactions.filter((tx) => {
      const haystack = [
        getField(tx, 'id', 'ID'),
        getField(tx, 'from', 'From', 'sender'),
        getField(tx, 'to', 'To', 'recipient'),
      ].join(' ').toLowerCase();
      return haystack.includes(normalizedQuery);
    });
  }, [transactions, normalizedQuery]);

  return (
    <section style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <div className="card" style={{
        borderRadius: 10,
        display: 'grid',
        gridTemplateColumns: '1fr auto',
        gap: 16,
        alignItems: 'center',
      }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexWrap: 'wrap', marginBottom: 6 }}>
            <h2 style={{ fontSize: '1.12rem', lineHeight: 1.2 }}>Mempool</h2>
            <span className="badge badge-amber">{transactions.length} pending</span>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.86rem' }}>
            Pending transactions waiting to be mined on <code className="mono">{nodeUrl}</code>.
          </p>
        </div>
        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', justifyContent: 'flex-end' }}>
          <button className="btn btn-ghost" type="button" onClick={onAction} title="Refresh mempool">
            Refresh
          </button>
          <button className="btn btn-primary" type="button" onClick={onGoToActions}>
            Mine in Actions
          </button>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'minmax(0, 1fr) auto', gap: 10, alignItems: 'center' }}>
        <input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search by sender, recipient, or transaction ID"
          aria-label="Search mempool transactions"
        />
        {query && (
          <button className="btn btn-ghost" type="button" onClick={() => setQuery('')}>
            Clear
          </button>
        )}
      </div>

      {transactions.length === 0 ? (
        <div className="card" style={{
          minHeight: 260,
          borderRadius: 10,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          textAlign: 'center',
          background: 'linear-gradient(180deg, #ffffff, #f8fafc)',
        }}>
          <div style={{ maxWidth: 460 }}>
            <div style={{
              width: 54,
              height: 54,
              borderRadius: 14,
              background: '#eff6ff',
              color: 'var(--accent-primary)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              margin: '0 auto 14px',
              fontWeight: 800,
              fontSize: '1.25rem',
            }}>
              0
            </div>
            <h3 style={{ fontSize: '1rem', marginBottom: 8 }}>No pending transactions</h3>
            <p style={{ color: 'var(--text-secondary)', fontSize: '0.88rem', marginBottom: 16 }}>
              Create a transfer or faucet request in Actions, then come back here to inspect it before mining.
            </p>
            <button className="btn btn-primary" type="button" onClick={onGoToActions}>
              Go to Actions
            </button>
          </div>
        </div>
      ) : filtered.length === 0 ? (
        <div className="card" style={{ borderRadius: 10, textAlign: 'center', color: 'var(--text-secondary)' }}>
          No transactions match "{query}".
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          {filtered.map((tx, index) => (
            <TransactionCard key={getField(tx, 'id', 'ID') || index} tx={tx} index={index} />
          ))}
        </div>
      )}
    </section>
  );
}
