import React from 'react';
import { StatCard } from './StatCard.jsx';
import { BarChart } from './BarChart.jsx';
import { Tooltip } from './Tooltip.jsx';
import { HEURISTIC_DEFINITIONS, OVERVIEW_DEFINITIONS } from './Definitions.jsx';

export const Dashboard = ({ data, onSelectBlock }) => {
  const s = data.analysis_summary;
  const fs = s.fee_rate_stats;

  return (
    <div className="app-container">
      <div className="overview-box">
        <h2 className="overview-title">
          <span className="overview-icon">📊</span>
          Analysis Overview
          <Tooltip id="dashboardSummary" type="overview" />
        </h2>
        <p className="overview-text">
          This analyzer examines Bitcoin blockchain data to detect transaction patterns and behaviors. 
          It identifies things like self-transfers, coinjoins, consolidation, and more - helping understand 
          what activities occurred in the blockchain.
        </p>
      </div>

      <h2 className="section-title">
        File Operations
        <Tooltip id="whatIsBlock" type="overview" />
      </h2>
      <div className="stats-grid">
        <StatCard label="Current File" value={data.file} sub={`${data.block_count} blocks analyzed`} />
        <StatCard 
          label="Total Transactions" 
          value={s.total_transactions_analyzed.toLocaleString()} 
          sub="All transactions analyzed"
        />
        <StatCard 
          label="Flagged Behaviors" 
          value={s.flagged_transactions.toLocaleString()} 
          sub={`${((s.flagged_transactions/s.total_transactions_analyzed)*100).toFixed(1)}% detection rate`}
        />
        <StatCard 
          label="Heuristics Applied" 
          value={s.heuristics_applied.length} 
          sub={s.heuristics_applied.slice(0,3).join(', ') + '...'} 
        />
      </div>

      <h2 className="section-title">
        Fee Rates (sat/vB)
        <Tooltip id="whatIsFee" type="overview" />
      </h2>
      <div className="stats-grid">
        <StatCard label="Minimum" value={fs.min_sat_vb.toFixed(1)} sub="Lowest fee in dataset" />
        <StatCard label="Median" value={fs.median_sat_vb.toFixed(1)} sub="Middle value" />
        <StatCard label="Average" value={fs.mean_sat_vb.toFixed(1)} sub="Mean fee" />
        <StatCard label="Maximum" value={fs.max_sat_vb.toFixed(1)} sub="Highest fee in dataset" />
      </div>

      <h2 className="section-title">
        Script Types
        <Tooltip id="p2wpkh" type="script" />
      </h2>
      <p className="section-subtitle">
        Different types of Bitcoin addresses used in outputs. Each represents a different way of securing Bitcoin.
      </p>
      <BarChart distribution={s.script_type_distribution} showTooltips={true} />

      <h2 className="section-title">
        Transaction Classifications
        <Tooltip id="simple_payment" type="classification" />
      </h2>
      <p className="section-subtitle">
        What type of activity each transaction represents - payments, transfers, consolidations, etc.
      </p>
      <BarChart distribution={s.classification_counts} showTooltips={true} isClassification={true} />

      <h2 className="section-title">
        Detected Heuristics
        <Tooltip id="whatIsHeuristic" type="overview" />
      </h2>
      <p className="section-subtitle">
        Patterns detected across all transactions. Each heuristic identifies specific behaviors.
      </p>
      <div className="heuristics-grid">
        {s.heuristics_applied.map(h => (
          <div key={h} className="heuristic-card">
            <div className="heuristic-name">
              {HEURISTIC_DEFINITIONS[h]?.name || h}
              <Tooltip id={h} type="heuristic" />
            </div>
            <div className="heuristic-count">
              {s.heuristic_counts?.[h]?.toLocaleString() || '0'}
            </div>
            <div className="heuristic-definition">
              {HEURISTIC_DEFINITIONS[h]?.definition}
            </div>
          </div>
        ))}
      </div>

      <h2 className="section-title">
        Block Explorer
        <Tooltip id="whatIsBlock" type="overview" />
      </h2>
      <p className="section-subtitle">
        Click on any block to see detailed transaction forensics and analysis.
      </p>
      <div className="blocks-grid">
        {data.blocks.map((blk, idx) => (
          <div key={blk.block_hash} className="block-card" onClick={() => onSelectBlock(idx)}>
            <div className="block-hash">{blk.block_hash}</div>
            <div className="block-height">Height {blk.block_height}</div>
            <div className="block-meta">
              <span>{blk.tx_count} TXs</span>
              <span className="text-danger">{blk.analysis_summary.flagged_transactions} Flagged</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
