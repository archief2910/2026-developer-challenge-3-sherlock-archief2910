import React from 'react';

export const BarChart = ({ distribution }) => {
  const entries = Object.entries(distribution || {}).sort((a,b) => b[1] - a[1]);
  if (!entries.length) return null;
  const max = entries[0][1];
  const colors = ['#818cf8', '#c084fc', '#34d399', '#fbbf24', '#f87171', '#60a5fa', '#f472b6'];

  return (
    <div className="bar-chart mt-4 mb-4">
      {entries.map(([type, count], i) => {
        const pct = (count / max * 100).toFixed(1);
        return (
          <div className="bar-row" key={type}>
            <span className="bar-label">{type}</span>
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
