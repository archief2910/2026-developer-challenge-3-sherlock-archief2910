import React from 'react';
import { Tooltip } from './Tooltip.jsx';
import { SCRIPT_TYPE_DEFINITIONS, CLASSIFICATION_DEFINITIONS } from './Definitions.jsx';

export const BarChart = ({ distribution, showTooltips = false, isClassification = false }) => {
  const entries = Object.entries(distribution || {}).sort((a,b) => b[1] - a[1]);
  if (!entries.length) return null;
  const max = entries[0][1];
  const colors = ['#818cf8', '#c084fc', '#34d399', '#fbbf24', '#f87171', '#60a5fa', '#f472b6'];

  const defs = isClassification ? CLASSIFICATION_DEFINITIONS : SCRIPT_TYPE_DEFINITIONS;

  return (
    <div className="bar-chart mt-4 mb-4">
      {entries.map(([type, count], i) => {
        const pct = (count / max * 100).toFixed(1);
        const def = defs[type];
        
        return (
          <div className="bar-row" key={type}>
            <span className="bar-label">
              {type}
              {showTooltips && def && <Tooltip id={type} type={isClassification ? 'classification' : 'script'} />}
            </span>
            <div className="bar-fill-container">
              <div 
                className="bar-fill" 
                data-count={count.toLocaleString()} 
                style={{ width: `${pct}%`, background: colors[i % colors.length] }} 
              />
            </div>
          </div>
        );
      })}
    </div>
  );
};
