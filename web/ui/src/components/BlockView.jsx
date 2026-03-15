import React from 'react';
import { StatCard } from './StatCard.jsx';
import { BarChart } from './BarChart.jsx';
import { TransactionGraph } from './TransactionGraph.jsx';
import { TransactionTable } from './TransactionTable.jsx';
import { Tooltip } from './Tooltip.jsx';

export const BlockView = ({ blockData, onBack }) => {
  const sum = blockData.analysis_summary;
  const fs = sum.fee_rate_stats;

  return (
    <div className="app-container">
      <button className="pill-button" onClick={onBack}>
        &larr; Back to File Overview
      </button>

      <h2 className="section-title">
        Transaction Flow
        <Tooltip id="whatIsTransaction" type="overview" />
      </h2>
      <TransactionGraph transactions={blockData.transactions} />

      <h2 className="section-title">
        Block #{blockData.block_height}
        <Tooltip id="whatIsBlock" type="overview" />
      </h2>
      <div className="stats-grid">
        <StatCard label="Block Hash" value={blockData.block_hash.slice(0, 10) + '...'} sub={blockData.block_hash} />
        <StatCard 
          label="Total Transactions" 
          value={blockData.tx_count.toLocaleString()} 
          sub="Transactions in this block" 
        />
        <StatCard 
          label="Flagged Transactions" 
          value={sum.flagged_transactions.toLocaleString()} 
          sub={`${((sum.flagged_transactions/blockData.tx_count)*100).toFixed(1)}% of block`} 
        />
        <StatCard 
          label="Block Time" 
          value={new Date(blockData.block_timestamp * 1000).toLocaleDateString()} 
          sub={new Date(blockData.block_timestamp * 1000).toLocaleTimeString()} 
        />
      </div>

      <h2 className="section-title">
        Fee Rates
        <Tooltip id="whatIsFee" type="overview" />
      </h2>
      <div className="stats-grid">
        <StatCard label="Minimum" value={fs.min_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Median" value={fs.median_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Average" value={fs.mean_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Maximum" value={fs.max_sat_vb.toFixed(1)} sub="sat/vB" />
      </div>

      <div style={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '2rem'}}>
        <div>
          <h2 className="section-title">
            Script Types
            <Tooltip id="p2wpkh" type="script" />
          </h2>
          <BarChart distribution={sum.script_type_distribution} showTooltips={true} />
        </div>
        <div>
          <h2 className="section-title">
            Classifications
            <Tooltip id="simple_payment" type="classification" />
          </h2>
          <BarChart distribution={sum.classification_counts} showTooltips={true} isClassification={true} />
        </div>
      </div>

      <h2 className="section-title">
        Transaction Forensics
        <Tooltip id="whatIsHeuristic" type="overview" />
      </h2>
      <p className="section-subtitle">
        Detailed analysis of each transaction in this block. Shows detected patterns and classifications.
      </p>
      <TransactionTable transactions={blockData.transactions} />
    </div>
  );
};
