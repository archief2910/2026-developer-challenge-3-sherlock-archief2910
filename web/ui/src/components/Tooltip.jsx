import React, { useState } from 'react';
import { 
  HEURISTIC_DEFINITIONS, 
  CLASSIFICATION_DEFINITIONS, 
  SCRIPT_TYPE_DEFINITIONS,
  OVERVIEW_DEFINITIONS 
} from './Definitions.jsx';

const style = {
  tooltip: {
    position: 'relative',
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: '18px',
    height: '18px',
    borderRadius: '50%',
    background: 'rgba(129, 140, 248, 0.2)',
    color: 'var(--accent)',
    fontSize: '11px',
    fontWeight: '700',
    cursor: 'help',
    marginLeft: '6px',
    transition: 'all 0.2s ease',
  },
  tooltipActive: {
    background: 'var(--accent)',
    color: '#fff',
  },
  tooltipBox: {
    position: 'absolute',
    top: '100%',
    left: '50%',
    transform: 'translateX(-50%)',
    marginTop: '10px',
    background: 'rgba(15, 23, 42, 0.98)',
    border: '1px solid var(--accent)',
    borderRadius: '12px',
    padding: '16px',
    minWidth: '320px',
    maxWidth: '420px',
    zIndex: 1000,
    boxShadow: '0 8px 32px rgba(0, 0, 0, 0.5), 0 0 20px var(--accent-glow)',
    animation: 'fadeIn 0.2s ease',
  },
  tooltipTitle: {
    color: '#fff',
    fontSize: '14px',
    fontWeight: '700',
    marginBottom: '8px',
    display: 'flex',
    alignItems: 'center',
    gap: '8px',
  },
  tooltipDefinition: {
    color: 'var(--text-muted)',
    fontSize: '13px',
    lineHeight: '1.5',
    marginBottom: '12px',
  },
  tooltipWhy: {
    color: 'var(--text-muted)',
    fontSize: '12px',
    lineHeight: '1.5',
    paddingTop: '10px',
    borderTop: '1px solid rgba(255,255,255,0.1)',
  },
  tooltipWhyTitle: {
    color: 'var(--accent)',
    fontWeight: '600',
    fontSize: '11px',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
    marginBottom: '4px',
  },
  tooltipExample: {
    color: 'var(--text-muted)',
    fontSize: '12px',
    lineHeight: '1.5',
    paddingTop: '10px',
    borderTop: '1px solid rgba(255,255,255,0.1)',
  },
  tooltipExampleTitle: {
    color: 'var(--success)',
    fontWeight: '600',
    fontSize: '11px',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
    marginBottom: '4px',
  }
};

const definitionsMap = {
  ...HEURISTIC_DEFINITIONS,
  ...CLASSIFICATION_DEFINITIONS,
  ...SCRIPT_TYPE_DEFINITIONS,
  ...OVERVIEW_DEFINITIONS
};

export const Tooltip = ({ id, type }) => {
  const [show, setShow] = useState(false);
  
  const def = definitionsMap[id];
  
  if (!def) return null;
  
  return (
    <span 
      style={{position: 'relative', display: 'inline-flex'}}
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
    >
      <span style={{...style.tooltip, ...(show ? style.tooltipActive : {})}}>?</span>
      {show && (
        <div style={style.tooltipBox}>
          <div style={style.tooltipTitle}>
            {def.name || def.title}
          </div>
          {def.definition && (
            <div style={style.tooltipDefinition}>{def.definition}</div>
          )}
          {def.content && (
            <div style={style.tooltipDefinition}>{def.content}</div>
          )}
          {def.whyImportant && (
            <div style={style.tooltipWhy}>
              <div style={style.tooltipWhyTitle}>Why It Matters</div>
              {def.whyImportant}
            </div>
          )}
          {def.example && (
            <div style={style.tooltipExample}>
              <div style={style.tooltipExampleTitle}>Example</div>
              {def.example}
            </div>
          )}
        </div>
      )}
    </span>
  );
};

export const HelpIcon = ({ text, style: customStyle = {} }) => {
  return (
    <span 
      style={{...style.tooltip, ...customStyle}}
      title={text}
    >
      ?
    </span>
  );
};

export default { Tooltip, HelpIcon };
