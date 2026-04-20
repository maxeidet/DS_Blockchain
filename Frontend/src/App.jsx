import React, { useState, useCallback } from 'react';
import { useNodeRegistry } from './hooks/useNodeRegistry';
import { useNodeData }     from './hooks/useNodeData';
import NodeCard            from './components/NodeCard';
import BlockList           from './components/BlockList';
import NodeActions         from './components/NodeActions';

// ─── Per-node data wrapper ────────────────────────────────────────────────────
function NodeDataProvider({ nodeUrl, children }) {
  const data = useNodeData(nodeUrl);
  return children(data);
}

// ─── App ─────────────────────────────────────────────────────────────────────
export default function App() {
  const { nodes, addNode, removeNode } = useNodeRegistry();
  const [selectedNode, setSelectedNode] = useState(null);
  const [newNodeUrl, setNewNodeUrl]     = useState('');
  const [activeTab, setActiveTab]       = useState('chain'); // 'chain' | 'actions'

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
        background: 'var(--bg-secondary)',
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <div style={{
            width: 36, height: 36,
            background: 'linear-gradient(135deg, var(--accent-primary), var(--accent-secondary))',
            borderRadius: 10,
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            fontSize: 18,
          }}>⛓</div>
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
          background: 'var(--bg-secondary)',
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
              placeholder="http://localhost:5001"
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
            <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--text-muted)' }}>
              Add a node to get started.
            </div>
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
                      {['chain', 'actions'].map(tab => (
                        <button
                          key={tab}
                          className={`btn ${activeTab === tab ? 'btn-primary' : 'btn-ghost'}`}
                          onClick={() => setActiveTab(tab)}
                        >
                          {tab === 'chain' ? '⛓ Chain' : '⚙ Actions'}
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
                  <div style={{ flex: 1, overflowY: 'auto', padding: '20px 24px' }}>
                    {activeTab === 'chain' ? (
                      <BlockList chain={data.chain} />
                    ) : (
                      <NodeActions nodeUrl={effective} onAction={data.refetch} />
                    )}
                  </div>
                </>
              )}
            </NodeDataProvider>
          )}
        </main>
      </div>
    </div>
  );
}
