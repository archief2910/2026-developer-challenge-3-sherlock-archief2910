import React from 'react';

export const TransactionGraph = ({ transactions }) => {
  if (!transactions || !transactions.length) return <p>No transactions to display.</p>;
  
  const coinjoins = transactions.filter(tx => tx.classification === 'coinjoin');
  const consolidations = transactions.filter(tx => tx.classification === 'consolidation');
  
  const visibleTxs = transactions.slice(0, Math.min(transactions.length, 250));

  return (
    <div className="tx-graph-container">
      <h3 className="section-title" style={{marginTop: 0}}>Transaction Flow Graph</h3>
      <div className="graph-legend">
        <span className="legend-item"><span className="legend-dot coinjoin"></span>CoinJoin ({coinjoins.length})</span>
        <span className="legend-item"><span className="legend-dot consolidation"></span>Consolidation ({consolidations.length})</span>
        <span className="legend-item"><span className="legend-dot batch_payment"></span>Batch Payment</span>
        <span className="legend-item"><span className="legend-dot self_transfer"></span>Self Transfer</span>
        <span className="legend-item"><span className="legend-dot normal"></span>Standard</span>
      </div>
      <div className="tx-timeline">
        {visibleTxs.map(tx => {
          let cls = 'normal';
          if (tx.classification === 'coinjoin') cls = 'coinjoin';
          else if (tx.classification === 'consolidation') cls = 'consolidation';
          else if (tx.classification === 'self_transfer') cls = 'self_transfer';
          else if (tx.classification === 'batch_payment') cls = 'batch_payment';
          
          return (
            <span 
              key={tx.txid}
              className={`tx-dot ${cls}`} 
              title={`${tx.txid.slice(0,12)}... (${tx.classification})`}
            ></span>
          );
        })}
        {transactions.length > 250 && <span className="text-muted ml-2">+{transactions.length - 250} more</span>}
      </div>
    </div>
  );
};
