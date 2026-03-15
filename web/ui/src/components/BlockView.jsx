import React from 'react';
import { StatCard } from './StatCard.jsx';
import { BarChart } from './BarChart.jsx';
import { TransactionGraph } from './TransactionGraph.jsx';
import { TransactionTable } from './TransactionTable.jsx';

export const BlockView = ({ blockData, onBack }) => {
  const sum = blockData.analysis_summary;
  const fs = sum.fee_rate_stats;

  return (
    <div className="app-container">
      <button className="pill-button" onClick={onBack}>
        &larr; Back to File Overview
      </button>

      <TransactionGraph transactions={blockData.transactions} />

      <h2 className="section-title">Block Summary: {blockData.block_height}</h2>
      <div className="stats-grid">
        <StatCard label="Block Hash" value={blockData.block_hash.slice(0, 10) + '...'} sub={blockData.block_hash} />
        <StatCard label="Total Transactions" value={blockData.tx_count.toLocaleString()} sub="Included in this block" />
        <StatCard label="Flagged Transactions" value={sum.flagged_transactions.toLocaleString()} sub={`${((sum.flagged_transactions/blockData.tx_count)*100).toFixed(1)}% of block`} />
        <StatCard label="Block Timestamp" value={new Date(blockData.block_timestamp * 1000).toLocaleString()} sub={`UNIX: ${blockData.block_timestamp}`} />
      </div>

      <h2 className="section-title">Fee Rates (sat/vB)</h2>
      <div className="stats-grid">
        <StatCard label="Minimum Fee" value={fs.min_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Median Fee" value={fs.median_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Mean Fee" value={fs.mean_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Maximum Fee" value={fs.max_sat_vb.toFixed(1)} sub="sat/vB" />
      </div>

      <div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem'}}>
        <div>
          <h2 className="section-title">Script Type Breakdown</h2>
          <BarChart distribution={sum.script_type_distribution} />
        </div>
        <div>
          <h2 className="section-title">Action Classifications</h2>
          <BarChart distribution={sum.classification_counts} />
        </div>
      </div>

      <h2 className="section-title">Transaction Forensics</h2>
      <TransactionTable transactions={blockData.transactions} />
    </div>
  );
};
