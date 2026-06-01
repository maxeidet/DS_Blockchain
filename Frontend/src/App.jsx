import React, { useState, useCallback } from 'react';
import { useNodeRegistry } from './hooks/useNodeRegistry';
import { useNodeData } from './hooks/useNodeData';
import NodeCard from './components/NodeCard';
import BlockChainView from './components/BlockList';
import NodeActions from './components/NodeActions';
import MempoolView from './components/MempoolView';
import AttackPanel from './components/AttackPanel';

// ─── Per-node data wrapper ────────────────────────────────────────────────────
function NodeDataProvider({ nodeUrl, children }) {
  const data = useNodeData(nodeUrl);
  return children(data);
}

// ─── Welcome Screen ────────────────────────────────────────────────────────────
function WelcomeScreen() {
  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', overflowY: 'auto', padding: '60px 40px', background: 'var(--bg-primary)' }}>
      <div style={{ maxWidth: 700, width: '100%', display: 'flex', flexDirection: 'column', gap: 24 }}>

        <div style={{ textAlign: 'center', marginBottom: 16 }}>
          <h1 style={{ fontSize: '2rem', fontWeight: 800, color: 'var(--text-primary)', marginBottom: 8, letterSpacing: '-0.03em' }}>Welcome to BlockView</h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '1.05rem' }}>Your control panel for the GO_Blockchain network.</p>
        </div>

        <div className="card" style={{ padding: '24px 32px' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: 12 }}>Quick Start</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: 16 }}>
            To use the explorer, you need to have at least one local blockchain node running. Open your terminal, navigate to the <code>GO_Blockchain</code> directory and run:
          </p>
          <pre style={{ background: '#1e293b', color: '#e2e8f0', padding: '16px', borderRadius: 8, fontFamily: 'var(--font-mono)', fontSize: '0.85rem', overflowX: 'auto', border: '1px solid var(--border)' }}>
            <code>go run . --api :8080 --p2p :6001 --advertise localhost:6001</code>
          </pre>
          <div style={{ marginTop: 16, background: 'var(--bg-muted)', padding: '12px 16px', borderRadius: 8, fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'flex', gap: 12 }}>
            <span style={{ fontSize: '1.2rem' }}></span>
            <div>
              Once the node is running, enter <strong>http://localhost:8080</strong> in the sidebar to the left and click <strong>+</strong> to connect!
            </div>
          </div>
        </div>

        <div className="card" style={{ padding: '24px 32px' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: 12 }}>Starting a P2P Network (UDP Discovery)</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', marginBottom: 16 }}>
            The nodes use automatic UDP discovery on port <strong>9999</strong> to find each other.
            You can start multiple nodes in separate terminal windows, and they will automatically connect and sync the chain!
          </p>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', marginBottom: 4 }}>Terminal 1 (Node 1)</div>
              <pre style={{ background: '#1e293b', color: '#e2e8f0', padding: '12px 16px', borderRadius: 8, fontFamily: 'var(--font-mono)', fontSize: '0.8rem', overflowX: 'auto' }}>
                <code>go run . --api :8080 --p2p :6001 --advertise localhost:6001</code>
              </pre>
            </div>
            <div>
              <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', marginBottom: 4 }}>Terminal 2 (Node 2)</div>
              <pre style={{ background: '#1e293b', color: '#e2e8f0', padding: '12px 16px', borderRadius: 8, fontFamily: 'var(--font-mono)', fontSize: '0.8rem', overflowX: 'auto' }}>
                <code>go run . --api :8081 --p2p :6002 --advertise localhost:6002</code>
              </pre>
            </div>
          </div>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.85rem', marginTop: 16 }}>
            <em>Then add both <code>http://localhost:8080</code> and <code>http://localhost:8081</code> in the sidebar to monitor them simultaneously.</em>
          </p>
        </div>

        <div className="card" style={{ padding: '24px 32px' }}>
          <h2 style={{ fontSize: '1.1rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: 12 }}>Security Lab</h2>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem' }}>
            Once you have added a node, navigate to the <strong>Security</strong> tab. There you can simulate 5 different known blockchain attacks against your node to test its cryptographic protection and validation logic.
          </p>
        </div>

      </div>
    </div>
  );
}

// ─── App ─────────────────────────────────────────────────────────────────────
export default function App() {
  const { nodes, addNode, removeNode } = useNodeRegistry();
  const [selectedNode, setSelectedNode] = useState(null);
  const [newNodeUrl, setNewNodeUrl] = useState('');
  const [activeTab, setActiveTab] = useState('chain'); // 'chain' | 'mempool' | 'actions' | 'security'

  const handleAddNode = useCallback((e) => {
    e.preventDefault();
    if (newNodeUrl.trim()) {
      addNode(newNodeUrl.trim());
      setNewNodeUrl('');
    }
  }, [newNodeUrl, addNode]);

  const effective = selectedNode ?? nodes[0] ?? null;

  return (
    <div style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Header */}
      <header style={{
        borderBottom: '1px solid var(--border)',
        padding: '16px 28px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        background: '#fff',
        boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <div style={{
            width: 36, height: 36,
            background: 'linear-gradient(135deg, var(--accent-primary), var(--accent-secondary))',
            borderRadius: 10,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 18,
          }}></div>
          <div>
            <div style={{ fontWeight: 700, fontSize: '1.05rem', letterSpacing: '-0.02em' }}>BlockView</div>
            <div style={{ fontSize: '0.72rem', color: 'var(--text-muted)' }}>Blockchain Explorer</div>
          </div>
        </div>
        <span className="badge badge-blue">{nodes.length} node{nodes.length !== 1 ? 's' : ''} tracked</span>
      </header>

      {/* Body */}
      <div style={{ flex: 1, display: 'grid', gridTemplateColumns: '320px 1fr', overflow: 'hidden' }}>

        {/* ── Left Sidebar ── */}
        <aside style={{
          borderRight: '1px solid var(--border)',
          padding: '20px 16px',
          overflowY: 'auto',
          background: '#fff',
          display: 'flex',
          flexDirection: 'column',
          gap: 12,
        }}>
          <h2 style={{ fontSize: '0.78rem', fontWeight: 600, color: 'var(--text-muted)', textTransform: 'uppercase', letterSpacing: '0.07em' }}>
            Nodes
          </h2>

          {/* Add node form */}
          <form onSubmit={handleAddNode} style={{ display: 'flex', gap: 8 }}>
            <input
              value={newNodeUrl}
              onChange={e => setNewNodeUrl(e.target.value)}
              placeholder="http://localhost:8080"
              style={{ flex: 1, fontSize: '0.8rem' }}
            />
            <button className="btn btn-primary" type="submit" style={{ padding: '8px 14px', flexShrink: 0 }}>+</button>
          </form>

          {/* Node cards */}
          {nodes.length === 0 && (
            <div style={{ color: 'var(--text-muted)', fontSize: '0.82rem', textAlign: 'center', padding: '20px 0' }}>
              Add a node URL above to get started.
            </div>
          )}
          {nodes.map(url => (
            <NodeDataProvider key={url} nodeUrl={url}>
              {(data) => (
                <NodeCard
                  nodeUrl={url}
                  data={data}
                  onRemove={removeNode}
                  selected={url === effective}
                  onClick={() => setSelectedNode(url)}
                />
              )}
            </NodeDataProvider>
          ))}
        </aside>

        {/* ── Main Content ── */}
        <main style={{ display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
          {!effective ? (
            <WelcomeScreen />
          ) : (
            <NodeDataProvider key={effective} nodeUrl={effective}>
              {(data) => (
                <>
                  {/* Subheader + tabs */}
                  <div style={{
                    borderBottom: '1px solid var(--border)',
                    padding: '14px 24px',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    flexShrink: 0,
                  }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                      <span className={`dot ${data.online ? 'dot-green' : 'dot-red'}`} />
                      <code className="mono" style={{ fontSize: '0.85rem' }}>{effective}</code>
                    </div>
                    <div style={{ display: 'flex', gap: 8 }}>
                      {[
                        { id: 'chain', label: 'Chain' },
                        { id: 'mempool', label: 'Mempool' },
                        { id: 'actions', label: 'Actions' },
                        { id: 'security', label: 'Security' },
                      ].map(({ id, label }) => (
                        <button
                          key={id}
                          className={`btn ${activeTab === id ? 'btn-primary' : 'btn-ghost'}`}
                          onClick={() => setActiveTab(id)}
                          style={id === 'security' ? {
                            borderColor: activeTab === 'security' ? undefined : 'rgba(220,38,38,0.3)',
                            color: activeTab === 'security' ? undefined : '#f87171',
                          } : {}}
                        >
                          {label}
                        </button>
                      ))}
                      <button
                        className="btn btn-ghost"
                        onClick={data.refetch}
                        title="Refresh"
                      >
                        ↻
                      </button>
                    </div>
                  </div>

                  {/* Tab content */}
                  {activeTab === 'chain' ? (
                    <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
                      <BlockChainView blocks={data.blocks} />
                    </div>
                  ) : activeTab === 'mempool' ? (
                    <div style={{ flex: 1, overflowY: 'auto', padding: '20px 24px' }}>
                      <MempoolView
                        mempool={data.mempool}
                        nodeUrl={effective}
                        onAction={data.refetch}
                        onGoToActions={() => setActiveTab('actions')}
                      />
                    </div>
                  ) : activeTab === 'actions' ? (
                    <div style={{ flex: 1, overflowY: 'auto', padding: '20px 24px' }}>
                      <NodeActions nodeUrl={effective} data={data} onAction={data.refetch} />
                    </div>
                  ) : (
                    <div style={{ flex: 1, overflowY: 'auto' }}>
                      <AttackPanel nodeUrl={effective} />
                    </div>
                  )}
                </>
              )}
            </NodeDataProvider>
          )}
        </main>
      </div>
    </div>
  );
}
