import React from 'react';

export const StatCard = ({ label, value, sub }) => (
  <div className="stat-card">
    <div className="stat-label">{label}</div>
    <div className="stat-value">{value}</div>
    <div className="stat-sub">{sub}</div>
  </div>
);
