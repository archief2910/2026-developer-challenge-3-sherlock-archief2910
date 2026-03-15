import React, { useState, useEffect } from 'react';
import { Dashboard } from './components/Dashboard.jsx';
import { BlockView } from './components/BlockView.jsx';
import { FileUpload } from './components/FileUpload.jsx';

const API = window.location.origin;

export const App = () => {
  const [files, setFiles] = useState([]);
  const [currentFile, setCurrentFile] = useState(null);
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(true);
  const [showUpload, setShowUpload] = useState(false);
  
  const [viewState, setViewState] = useState({ type: 'dashboard', val: null });

  useEffect(() => {
    fetch(`${API}/api/blocks`)
      .then(r => r.json())
      .then(d => {
        if (d.ok && d.files && d.files.length) {
          setFiles(d.files);
          loadFileData(d.files[0]);
        } else {
          setShowUpload(true);
          setLoading(false);
        }
      })
      .catch(e => {
        setError(e.message);
        setLoading(false);
      });
  }, []);

  const loadFileData = (filename) => {
    setLoading(true);
    setCurrentFile(filename);
    setViewState({ type: 'dashboard', val: null });
    setShowUpload(false);
    
    fetch(`${API}/api/blocks/${filename}`)
      .then(r => r.json())
      .then(d => {
        setData(d);
        setLoading(false);
      })
      .catch(e => {
        setError(e.message);
        setLoading(false);
      });
  };

  const handleUploadComplete = (filename, result) => {
    setFiles(prev => {
      const newFiles = [...prev, filename];
      return newFiles;
    });
    setData(result);
    setCurrentFile(filename);
    setShowUpload(false);
    setError(null);
  };

  return (
    <>
      <header className="top-bar">
        <div className="logo">
          <span className="logo-icon">🕵️</span>
          <h1>Sherlock</h1>
          <span className="badge">React Chain Analyzer</span>
        </div>
        <div className="nav-right">
          <nav className="nav-links">
            {files.map(f => (
              <button 
                key={f} 
                className={`nav-btn ${f === currentFile ? 'active' : ''}`}
                onClick={() => loadFileData(f)}
              >
                {f}
              </button>
            ))}
          </nav>
          <button className="upload-nav-btn" onClick={() => setShowUpload(true)}>
            + Upload
          </button>
        </div>
      </header>

      <main>
        {loading && (
          <div className="loading-screen">
            <div className="spinner"></div>
            <h3 style={{color: '#fff'}}>Synthesizing On-Chain Forensics...</h3>
          </div>
        )}

        {showUpload && !loading && (
          <FileUpload onUploadComplete={handleUploadComplete} />
        )}
        
        {error && !loading && !showUpload && (
          <div className="app-container" style={{textAlign: 'center', marginTop: '4rem'}}>
            <div style={{color: 'var(--danger)', fontSize: '1.25rem', marginBottom: '1rem'}}>
              Server Error Details
            </div>
            <code>{error}</code>
          </div>
        )}

        {!loading && data && viewState.type === 'dashboard' && !showUpload && (
          <Dashboard data={data} onSelectBlock={(idx) => setViewState({ type: 'block', val: idx })} />
        )}

        {!loading && data && viewState.type === 'block' && !showUpload && (
          <BlockView blockData={data.blocks[viewState.val]} onBack={() => setViewState({ type: 'dashboard' })} />
        )}
      </main>
    </>
  );
};
