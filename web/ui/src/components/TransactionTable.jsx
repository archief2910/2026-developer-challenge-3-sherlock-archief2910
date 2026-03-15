import React, { useState, useMemo } from 'react';

export const TransactionTable = ({ transactions }) => {
  const [filterClass, setFilterClass] = useState('all');
  const [filterHeur, setFilterHeur] = useState('all');
  const [expandedTx, setExpandedTx] = useState(null);

  const classifications = ['all', 'simple_payment', 'consolidation', 'coinjoin', 'self_transfer', 'batch_payment', 'unknown'];
  const heuristics = ['all', 'address_reuse', 'change_detection', 'cioh', 'coinjoin', 'consolidation', 'op_return', 'peeling_chain', 'round_number_payment', 'self_transfer'];

  const filtered = useMemo(() => {
    let res = transactions || [];
    if (filterClass !== 'all') res = res.filter(t => t.classification === filterClass);
    if (filterHeur !== 'all') res = res.filter(t => t.heuristics && t.heuristics[filterHeur] && t.heuristics[filterHeur].detected);
    return res;
  }, [transactions, filterClass, filterHeur]);

  const toggleExpand = (txid) => {
    setExpandedTx(expandedTx === txid ? null : txid);
  };

  const visibleCount = Math.min(filtered.length, 100);

  return (
    <div className="data-table-container">
      <div className="filter-bar">
        <div className="filter-group">
          <label>Classification:</label>
          <select className="filter-select" value={filterClass} onChange={e => setFilterClass(e.target.value)}>
            {classifications.map(c => <option key={c} value={c}>{c === 'all' ? 'All' : c.replace('_', ' ')}</option>)}
          </select>
        </div>
        <div className="filter-group">
          <label>Heuristic:</label>
          <select className="filter-select" value={filterHeur} onChange={e => setFilterHeur(e.target.value)}>
            {heuristics.map(h => <option key={h} value={h}>{h === 'all' ? 'All' : h.replace('_', ' ')}</option>)}
          </select>
        </div>
        <span style={{marginLeft: 'auto', fontSize: '0.85rem', color: 'var(--text-muted)'}}>
          Showing {visibleCount} of {filtered.length} (Total: {transactions?.length || 0})
        </span>
      </div>

      <table className="data-table">
        <thead>
          <tr>
            <th>TxID</th>
            <th>Classification</th>
            <th>Triggered Heuristics</th>
            <th style={{textAlign: 'right'}}>Details</th>
          </tr>
        </thead>
        <tbody>
          {filtered.slice(0, visibleCount).map(tx => {
            const detectedKeys = Object.keys(tx.heuristics).filter(k => tx.heuristics[k].detected);
            const isExpanded = expandedTx === tx.txid;
            
            return (
              <React.Fragment key={tx.txid}>
                <tr className="tx-row" onClick={() => toggleExpand(tx.txid)}>
                  <td><span className="tx-hash">{tx.txid.slice(0,12)}...{tx.txid.slice(-8)}</span></td>
                  <td><span className={`badge-sm badge-${tx.classification}`}>{tx.classification.replace('_', ' ')}</span></td>
                  <td>
                    {detectedKeys.length > 0 ? detectedKeys.map(h => <span key={h} className="heuristic-tag detected">{h}</span>) : <span className="heuristic-tag">none</span>}
                  </td>
                  <td style={{textAlign: 'right'}}><span className="text-muted">{isExpanded ? '▼' : '▶'}</span></td>
                </tr>
                {isExpanded && (
                  <tr className="tx-details-row">
                    <td colSpan="4">
                      <div className="tx-details">
                        <div style={{marginBottom: '1rem'}}>
                          <strong>Transaction ID:</strong> <code style={{color: 'var(--accent)'}}>{tx.txid}</code>
                        </div>
                        <div className="details-grid">
                          {Object.entries(tx.heuristics).map(([hName, res]) => (
                            <div key={hName} className="detail-box">
                              <h4>{hName.replace('_', ' ')}</h4>
                              {res.detected ? (
                                <div>
                                  <div className="text-danger" style={{fontWeight: 700, marginBottom: '0.25rem'}}>DETECTED</div>
                                  {res.confidence && <div><span className="text-muted">Confidence:</span> {res.confidence}</div>}
                                  {res.method && <div><span className="text-muted">Method:</span> {res.method}</div>}
                                  {res.likely_change_index !== undefined && <div><span className="text-muted">Change Index:</span> {res.likely_change_index}</div>}
                                  {res.protocol && <div><span className="text-muted">Data Protocol:</span> {res.protocol}</div>}
                                  {res.cross_tx && <div><span className="text-muted">Cross-TX Link:</span> Yes</div>}
                                </div>
                              ) : (
                                <div className="text-muted">Not detected</div>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    </td>
                  </tr>
                )}
              </React.Fragment>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};
