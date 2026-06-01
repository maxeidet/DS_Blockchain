import React, { useState, useCallback, useEffect } from 'react';
import { runAttackTest } from '../api/blockchain';

// ─── Static attack metadata (steps + defense info per attack name) ────────────
const ATTACK_META = {
  'Replay Attack': {
    icon: '',
    tagline: 'Send the same transaction twice',
    vector: 'The attacker intercepts a valid, already approved transaction and resends it to steal funds a second time.',
    steps: [
      { label: 'Create valid tx', detail: 'Alice → Bob, 10 coins, nonce=N, signed with Alice\'s private key' },
      { label: 'First submission', detail: 'The transaction is accepted and added to the mempool ' },
      { label: 'Replay attempt', detail: 'The exact same tx object is sent again — same ID, same nonce' },
      { label: 'System check', detail: 'AddTransactionToMempool() → seenTxIDs[tx.ID] == true → REJECT' },
    ],
    defense: 'Nonce ordering + Duplicate TX-ID',
    defenseFile: 'blockchain.go → AddTransactionToMempool()',
    defenseCode: `added := bc.Mempool.Add(tx)
if !added {
    return fmt.Errorf("transaction already exists in mempool: %s", tx.ID)
}`,
    color: '#7c3aed',
  },
  'Invalid Signature': {
    icon: '',
    tagline: 'Forge a digital signature',
    vector: 'The attacker modifies the transaction payload or signature bytes to send coins without true authorization.',
    steps: [
      { label: 'Create valid tx', detail: 'Alice → Bob, 10 coins, valid ECDSA signature' },
      { label: 'Corruption', detail: 'All bytes in the signature are XORed with 0xFF — a bitflip on the entire signature' },
      { label: 'Submission', detail: 'tx.Signature = corrupted hex string, a new tx.ID is calculated' },
      { label: 'System check', detail: 'VerifySignature() → ecdsa.Verify() returns false → REJECT' },
    ],
    defense: 'ECDSA P-256 signature verification',
    defenseFile: 'blockchain.go → validateTransferBasics()',
    defenseCode: `ok, err := VerifySignature(pubKeyBytes, signBytes, tx.Signature)
if err != nil {
    return fmt.Errorf("signature verification failed: %w", err)
}
if !ok {
    return errors.New("invalid signature")
}`,
    color: '#dc2626',
  },
  'Wrong Public Key': {
    icon: '',
    tagline: 'Use wrong public key to hijack identity',
    vector: 'The attacker signs with Alice\'s private key but claims the tx comes from Eve (via Eve\'s public key and address).',
    steps: [
      { label: 'Build tx with Eve\'s ID', detail: 'tx.From = Eve.Address, tx.PublicKey = Eve.PublicKey' },
      { label: 'Sign with Alice\'s key', detail: 'SignBytes(alice.PrivateKey, signBytes) — wrong key used' },
      { label: 'Submission', detail: 'The transaction looks valid cryptographically but the address doesn\'t match' },
      { label: 'System check', detail: 'AddressFromPublicKey(pubKeyBytes) ≠ tx.From → REJECT' },
    ],
    defense: 'Address derivation from public key',
    defenseFile: 'blockchain.go → validateTransferBasics()',
    defenseCode: `derivedAddress := AddressFromPublicKey(pubKeyBytes)
if derivedAddress != tx.From {
    return errors.New("from address does not match public key")
}`,
    color: '#d97706',
  },
  'Overdraft': {
    icon: '',
    tagline: 'Spend more coins than owned',
    vector: 'The attacker tries to send 9999 coins more than Alice\'s actual balance to create money out of thin air.',
    steps: [
      { label: 'Check balance', detail: 'Alice\'s balance = GetBalance(alice.Address) → e.g. 200 coins' },
      { label: 'Create overdraft tx', detail: 'Alice → Bob, amount = balance + 9999 — signed with valid key' },
      { label: 'Submission', detail: 'The transaction is cryptographically valid but economically impossible' },
      { label: 'System check', detail: 'tempBalances[tx.From] < tx.Amount → "insufficient balance" → REJECT' },
    ],
    defense: 'State-based balance validation',
    defenseFile: 'blockchain.go → ValidateBlockTransactionsAgainstState()',
    defenseCode: `if tempBalances[tx.From] < tx.Amount {
    return fmt.Errorf(
        "address %s trying to spend %d but only has %d",
        tx.From, tx.Amount, tempBalances[tx.From],
    )
}`,
    color: '#16a34a',
  },
  'Future Timestamp': {
    icon: '',
    tagline: 'Manipulate block time to fool the network',
    vector: 'The attacker sets a timestamp 3 hours into the future on a block, which can manipulate difficulty adjustments.',
    steps: [
      { label: 'Build block', detail: 'block.Timestamp = time.Now().Add(3h).UnixMilli()' },
      { label: 'Mine the block', detail: 'MineBlock() finds a valid nonce — PoW requirement is met' },
      { label: 'Submission', detail: 'IsBlockValid(badBlock, prev) is called during block acceptance' },
      { label: 'System check', detail: 'block.Timestamp > now + 2h → "timestamp too far in the future" → REJECT' },
    ],
    defense: 'Timestamp validation (2h limit)',
    defenseFile: 'blockchain.go → IsBlockValid()',
    defenseCode: `if newBlock.Timestamp > time.Now().Add(2*time.Hour).UnixMilli() {
    return errors.New("block timestamp is too far in the future")
}`,
    color: '#2563eb',
  },
};

const DEFAULT_ORDER = [
  'Replay Attack',
  'Invalid Signature',
  'Wrong Public Key',
  'Overdraft',
  'Future Timestamp',
];

// ─── Sub-components ───────────────────────────────────────────────────────────

function Modal({ isOpen, onClose, title, children }) {
  useEffect(() => {
    if (isOpen) document.body.style.overflow = 'hidden';
    else document.body.style.overflow = 'unset';
    return () => { document.body.style.overflow = 'unset'; };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div style={{
      position: 'fixed', top: 0, left: 0, width: '100vw', height: '100vh',
      background: 'rgba(15, 23, 42, 0.6)', backdropFilter: 'blur(4px)',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      zIndex: 9999, padding: 20
    }} onClick={onClose}>
      <div style={{
        background: 'var(--bg-secondary)', borderRadius: 16, width: '100%', maxWidth: 650,
        boxShadow: 'var(--shadow-lg)', overflow: 'hidden', animation: 'fadeIn 0.2s ease',
        display: 'flex', flexDirection: 'column', maxHeight: '90vh'
      }} onClick={e => e.stopPropagation()}>
        <div style={{
          padding: '20px 24px', borderBottom: '1px solid var(--border)',
          display: 'flex', justifyContent: 'space-between', alignItems: 'center'
        }}>
          <h3 style={{ fontSize: '1.2rem', fontWeight: 700, color: 'var(--text-primary)' }}>{title}</h3>
          <button onClick={onClose} style={{
            background: 'var(--bg-muted)', border: 'none', width: 32, height: 32,
            borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center',
            cursor: 'pointer', color: 'var(--text-secondary)'
          }}></button>
        </div>
        <div style={{ padding: '24px', overflowY: 'auto' }}>
          {children}
        </div>
      </div>
    </div>
  );
}


function StepFlow({ steps, running, result }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
      {steps.map((step, i) => {
        const isLast = i === steps.length - 1;
        const stepDone = result !== null || !running;
        const isDefenseStep = isLast;

        let stepColor = 'var(--text-muted)';
        if (result !== null) {
          stepColor = isDefenseStep
            ? (result.blocked ? 'var(--accent-success)' : 'var(--accent-danger)')
            : 'var(--text-secondary)';
        }

        return (
          <div key={i} style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
            {/* Connector line + dot */}
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', flexShrink: 0 }}>
              <div style={{
                width: 26, height: 26,
                borderRadius: '50%',
                background: isDefenseStep && result !== null
                  ? (result.blocked ? 'var(--accent-success)' : 'var(--accent-danger)')
                  : 'var(--bg-muted)',
                border: `2px solid ${stepColor}`,
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: 11, fontWeight: 700, 
                color: isDefenseStep && result !== null ? '#fff' : stepColor,
                flexShrink: 0,
                transition: 'all 0.3s',
              }}>
                {isDefenseStep && result !== null
                  ? (result.blocked ? '' : '✗')
                  : i + 1}
              </div>
              {!isLast && (
                <div style={{ width: 2, height: 28, background: 'var(--border)', flexShrink: 0 }} />
              )}
            </div>

            {/* Step content */}
            <div style={{ paddingBottom: isLast ? 0 : 28, minWidth: 0 }}>
              <div style={{
                fontSize: '0.78rem', fontWeight: 600,
                color: isDefenseStep && result !== null
                  ? (result.blocked ? 'var(--accent-success)' : 'var(--accent-danger)')
                  : 'var(--text-primary)',
                marginBottom: 2,
              }}>
                {step.label}
              </div>
              <div style={{ fontSize: '0.72rem', color: 'var(--text-secondary)', lineHeight: 1.4 }}>
                {step.detail}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function ResultBadge({ blocked }) {
  if (blocked === null) return null;
  return (
    <div style={{
      display: 'inline-flex', alignItems: 'center', gap: 6,
      padding: '4px 12px',
      borderRadius: 999,
      fontWeight: 700, fontSize: '0.75rem', letterSpacing: '0.06em',
      background: blocked ? 'rgba(22,163,74,0.1)' : 'rgba(220,38,38,0.1)',
      color: blocked ? 'var(--accent-success)' : 'var(--accent-danger)',
      border: `1px solid ${blocked ? 'rgba(22,163,74,0.3)' : 'rgba(220,38,38,0.3)'}`,
    }}>
      {blocked ? 'BLOCKED' : 'LEAKED'}
    </div>
  );
}

function AttackCard({ name, result, running, onOpenInfo }) {
  const meta = ATTACK_META[name] || {};
  const isRunning = running === name;
  const cardResult = result?.[name] ?? null;

  return (
    <div className="card" style={{
      border: `1px solid ${cardResult !== null
        ? (cardResult.blocked ? 'var(--accent-success)' : 'var(--accent-danger)')
        : 'var(--border)'}`,
      padding: 0,
      display: 'flex', flexDirection: 'column',
      boxShadow: cardResult !== null
        ? (cardResult.blocked
          ? '0 0 15px rgba(22,163,74,0.1)'
          : '0 0 15px rgba(220,38,38,0.1)')
        : 'var(--shadow-sm)',
    }}>

      {/* Card header */}
      <div style={{
        padding: '16px 20px',
        borderBottom: '1px solid var(--border)',
        display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 12,
        background: 'var(--bg-secondary)',
        borderTopLeftRadius: 12, borderTopRightRadius: 12,
      }}>
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
          <div style={{
            width: 40, height: 40, borderRadius: 10, flexShrink: 0,
            background: `${meta.color}15`,
            border: `1px solid ${meta.color}30`,
            display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18,
          }}>
            {meta.icon}
          </div>
          <div>
            <div style={{ fontWeight: 700, color: 'var(--text-primary)', fontSize: '0.9rem' }}>{name}</div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.72rem', marginTop: 2 }}>{meta.tagline}</div>
          </div>
        </div>
        <ResultBadge blocked={cardResult?.blocked ?? null} />
      </div>

      {/* Body */}
      <div style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 16 }}>

        {/* Step-by-step flow */}
        <div>
          <div style={{ fontSize: '0.68rem', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.08em', marginBottom: 10 }}>
            Execution Flow
          </div>
          {isRunning ? (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, color: 'var(--text-secondary)', fontSize: '0.8rem', padding: '12px 0' }}>
              <div style={{
                width: 16, height: 16,
                border: '2px solid var(--border)',
                borderTopColor: meta.color,
                borderRadius: '50%',
                animation: 'spin 0.7s linear infinite',
              }} />
              Running attack simulation...
            </div>
          ) : (
            <StepFlow steps={meta.steps ?? []} running={isRunning} result={cardResult} />
          )}
        </div>

        {/* Result area — shown after run */}
        {cardResult !== null && (
          <div style={{
            background: 'var(--bg-secondary)', borderRadius: 8, overflow: 'hidden',
            border: `1px solid ${cardResult.blocked ? 'rgba(22,163,74,0.4)' : 'rgba(220,38,38,0.4)'}`,
            animation: 'fadeIn 0.3s ease',
          }}>
            <div style={{ padding: '10px 14px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <div>
                <div style={{ fontSize: '0.65rem', color: 'var(--text-muted)', marginBottom: 4 }}>
                  {cardResult.blocked ? 'System Error Message:' : 'System Response:'}
                </div>
                <code style={{
                  fontSize: '0.75rem', fontFamily: 'var(--font-mono)',
                  color: cardResult.blocked ? 'var(--accent-success)' : 'var(--accent-danger)',
                  wordBreak: 'break-all', display: 'block',
                }}>
                  {cardResult.error || '(no error — attack was accepted)'}
                </code>
              </div>
              <button onClick={() => onOpenInfo(name)} className="btn btn-ghost" style={{ padding: '4px 8px', fontSize: '0.75rem' }}>
                Info
              </button>
            </div>
          </div>
        )}

        {/* Expected (always shown) */}
        {cardResult === null && (
          <div style={{ fontSize: '0.72rem', color: 'var(--text-muted)' }}>
            <span>Expected: </span>
            {meta.steps?.[meta.steps.length - 1]?.detail ?? '—'}
          </div>
        )}
      </div>
    </div>
  );
}

function SummaryBar({ summary }) {
  if (!summary) return null;
  const pct = Math.round((summary.blocked / summary.total) * 100);
  return (
    <div className="card" style={{
      display: 'flex', alignItems: 'center', gap: 20,
      flexWrap: 'wrap',
    }}>
      <div style={{ display: 'flex', gap: 20, flex: 1 }}>
        <Stat label="Total" value={summary.total} color="var(--text-primary)" />
        <Stat label="Blocked" value={summary.blocked} color="var(--accent-success)" />
        <Stat label="Leaked" value={summary.leaked} color={summary.leaked > 0 ? 'var(--accent-danger)' : 'var(--text-muted)'} />
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4, minWidth: 160 }}>
        <div style={{ fontSize: '0.7rem', color: 'var(--text-secondary)', display: 'flex', justifyContent: 'space-between' }}>
          <span>Protection Rate</span>
          <span style={{ color: pct === 100 ? 'var(--accent-success)' : 'var(--accent-warning)', fontWeight: 600 }}>{pct}%</span>
        </div>
        <div style={{ height: 6, background: 'var(--bg-muted)', borderRadius: 999, overflow: 'hidden' }}>
          <div style={{
            height: '100%', width: `${pct}%`,
            background: pct === 100 ? 'var(--accent-success)' : 'var(--accent-warning)',
            borderRadius: 999, transition: 'width 0.6s ease',
          }} />
        </div>
      </div>
    </div>
  );
}

function Stat({ label, value, color }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
      <div style={{ fontSize: '1.4rem', fontWeight: 700, color, lineHeight: 1 }}>{value}</div>
      <div style={{ fontSize: '0.68rem', color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.06em' }}>{label}</div>
    </div>
  );
}

// ─── Main component ───────────────────────────────────────────────────────────
export default function AttackPanel({ nodeUrl }) {
  const [results, setResults] = useState(null);   // { [name]: AttackResult }
  const [summary, setSummary] = useState(null);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState(null);
  const [activeModal, setActiveModal] = useState(null);

  const handleRunAll = useCallback(async () => {
    setRunning(true);
    setError(null);
    setResults(null);
    setSummary(null);
    try {
      const data = await runAttackTest(nodeUrl);
      const byName = {};
      for (const r of data.attacks) byName[r.name] = r;
      setResults(byName);
      setSummary(data.summary);
    } catch (err) {
      setError(err.message);
    } finally {
      setRunning(false);
    }
  }, [nodeUrl]);

  const handleReset = () => {
    setResults(null);
    setSummary(null);
    setError(null);
  };

  const selectedMeta = activeModal ? ATTACK_META[activeModal] : null;

  return (
    <div style={{
      padding: '24px',
      display: 'flex', flexDirection: 'column', gap: 20,
    }}>
      {/* Header */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
        <div>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: 4, display: 'flex', alignItems: 'center', gap: 10 }}>
            <span style={{
              background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.2)',
              borderRadius: 8, padding: '4px 10px', fontSize: '0.8rem', color: 'var(--accent-danger)',
            }}>
              Security Lab
            </span>
            Attack Resistance Test
          </h2>
          <p style={{ fontSize: '0.78rem', color: 'var(--text-secondary)', maxWidth: 520 }}>
            Simulates known blockchain attacks against an isolated chain instance and shows exactly which defense mechanism catches each attack.
          </p>
        </div>

        <div style={{ display: 'flex', gap: 8 }}>
          {results && (
            <button
              onClick={handleReset}
              className="btn btn-ghost"
              style={{ fontSize: '0.82rem' }}
            >
              Reset
            </button>
          )}
          <button
            onClick={handleRunAll}
            disabled={running}
            className="btn btn-primary"
            style={{
              padding: '10px 20px',
              fontWeight: 600, fontSize: '0.85rem',
              opacity: running ? 0.7 : 1,
            }}
          >
            {running ? (
              <>
                <div style={{
                  width: 14, height: 14,
                  border: '2px solid rgba(255,255,255,0.4)',
                  borderTopColor: '#fff',
                  borderRadius: '50%',
                  animation: 'spin 0.7s linear infinite',
                }} />
                Running tests...
              </>
            ) : (
              'Run all attacks'
            )}
          </button>
        </div>
      </div>

      {/* Error banner */}
      {error && (
        <div style={{
          background: 'rgba(220,38,38,0.05)', border: '1px solid rgba(220,38,38,0.2)',
          borderRadius: 10, padding: '12px 16px',
          color: 'var(--accent-danger)', fontSize: '0.82rem',
        }}>
          Error connecting to node: {error}
          <div style={{ marginTop: 4, color: 'var(--text-muted)', fontSize: '0.72rem' }}>
            Ensure the node is running and accessible at {nodeUrl}
          </div>
        </div>
      )}

      {/* Summary bar */}
      {summary && <SummaryBar summary={summary} />}

      {/* Attack cards grid */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(420px, 1fr))',
        gap: 16,
      }}>
        {DEFAULT_ORDER.map(name => (
          <AttackCard
            key={name}
            name={name}
            result={results}
            running={running ? name : null}
            onOpenInfo={setActiveModal}
          />
        ))}
      </div>

      {/* Disclaimer */}
      <div style={{
        fontSize: '0.68rem', color: 'var(--text-muted)',
        textAlign: 'center', padding: '8px 0',
      }}>
        All attacks are run against an isolated blockchain instance — the active chain on the node is not affected.
      </div>

      {/* Info Modal */}
      <Modal 
        isOpen={!!activeModal} 
        onClose={() => setActiveModal(null)} 
        title={activeModal ? `${selectedMeta.icon} ${activeModal} Details` : ''}
      >
        {selectedMeta && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
            <div>
              <h4 style={{ fontSize: '0.85rem', color: 'var(--text-primary)', marginBottom: 6 }}>Attack Vector</h4>
              <div style={{
                background: 'var(--bg-muted)', borderRadius: 8, padding: '12px 16px',
                fontSize: '0.85rem', color: 'var(--text-secondary)', lineHeight: 1.5,
                borderLeft: `4px solid ${selectedMeta.color}`
              }}>
                {selectedMeta.vector}
              </div>
            </div>
            
            <div>
              <h4 style={{ fontSize: '0.85rem', color: 'var(--text-primary)', marginBottom: 6 }}>How it is prevented</h4>
              <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: 12 }}>
                The blockchain protects against this using <strong>{selectedMeta.defense}</strong>. 
                This logic is located in <code>{selectedMeta.defenseFile}</code>.
              </div>
              
              <div style={{ position: 'relative' }}>
                <div style={{ 
                  position: 'absolute', top: -10, left: 16, 
                  background: 'var(--bg-primary)', padding: '0 8px', 
                  fontSize: '0.7rem', fontWeight: 600, color: 'var(--accent-primary)',
                  border: '1px solid var(--border)', borderRadius: 4
                }}>
                  Defense Implementation
                </div>
                <pre style={{
                  background: '#1e293b', color: '#e2e8f0', padding: '20px 16px 16px 16px',
                  borderRadius: 8, fontSize: '0.8rem', fontFamily: 'var(--font-mono)',
                  overflowX: 'auto', border: '1px solid var(--border)', margin: 0,
                  boxShadow: 'inset 0 2px 4px rgba(0,0,0,0.1)'
                }}>
                  {selectedMeta.defenseCode}
                </pre>
              </div>
            </div>
          </div>
        )}
      </Modal>

    </div>
  );
}
