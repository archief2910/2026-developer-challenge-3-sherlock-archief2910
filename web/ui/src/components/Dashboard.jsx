import React from 'react';
import { StatCard } from './StatCard.jsx';
import { BarChart } from './BarChart.jsx';

export const Dashboard = ({ data, onSelectBlock }) => {
  const s = data.analysis_summary;
  const fs = s.fee_rate_stats;

  return (
    <div className="app-container">
      <h2 className="section-title">File Operations</h2>
      <div className="stats-grid">
        <StatCard label="Current File" value={data.file} sub={`${data.block_count} blocks analyzed`} />
        <StatCard label="Total Transactions" value={s.total_transactions_analyzed.toLocaleString()} sub="Global analysis pool" />
        <StatCard label="Flagged Behaviors" value={s.flagged_transactions.toLocaleString()} sub={`${((s.flagged_transactions/s.total_transactions_analyzed)*100).toFixed(1)}% detection rate`} />
        <StatCard label="Confidence Heuristics" value={s.heuristics_applied.length} sub={s.heuristics_applied.slice(0,3).join(', ') + '...'} />
      </div>

      <h2 className="section-title">Global Fee Distribution</h2>
      <div className="stats-grid">
        <StatCard label="Dataset Min" value={fs.min_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Dataset Median" value={fs.median_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Dataset Mean" value={fs.mean_sat_vb.toFixed(1)} sub="sat/vB" />
        <StatCard label="Dataset Max" value={fs.max_sat_vb.toFixed(1)} sub="sat/vB" />
      </div>

      <h2 className="section-title">Script Archetypes</h2>
      <BarChart distribution={s.script_type_distribution} />

      <h2 className="section-title">Block Explorer</h2>
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
