import React, { useState, useRef } from 'react';

export const FileUpload = ({ onUploadComplete }) => {
  const [files, setFiles] = useState({ blk: null, rev: null, xor: null });
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState(null);
  const [dragOver, setDragOver] = useState({ blk: false, rev: false, xor: false });

  const handleFileChange = (type, e) => {
    const file = e.target.files[0];
    if (file) {
      setFiles(prev => ({ ...prev, [type]: file }));
      setError(null);
    }
  };

  const handleDragOver = (e, type) => {
    e.preventDefault();
    setDragOver(prev => ({ ...prev, [type]: true }));
  };

  const handleDragLeave = (e, type) => {
    e.preventDefault();
    setDragOver(prev => ({ ...prev, [type]: false }));
  };

  const handleDrop = (e, type) => {
    e.preventDefault();
    setDragOver(prev => ({ ...prev, [type]: false }));
    const file = e.dataTransfer.files[0];
    if (file) {
      setFiles(prev => ({ ...prev, [type]: file }));
      setError(null);
    }
  };

  const handleUpload = async () => {
    if (!files.blk || !files.rev || !files.xor) {
      setError('Please select all three files: blk.dat, rev.dat, and xor.dat');
      return;
    }

    setUploading(true);
    setError(null);

    const formData = new FormData();
    formData.append('blk', files.blk);
    formData.append('rev', files.rev);
    formData.append('xor', files.xor);

    try {
      const response = await fetch(`${window.location.origin}/api/upload`, {
        method: 'POST',
        body: formData,
      });

      const data = await response.json();

      if (!response.ok || !data.ok) {
        throw new Error(data.error?.message || 'Upload failed');
      }

      onUploadComplete(data.file, data.result);
    } catch (err) {
      setError(err.message);
    } finally {
      setUploading(false);
    }
  };

  const renderDropZone = (type, label, accept) => {
    const isDragOver = dragOver[type];
    const hasFile = files[type];

    return (
      <div
        className={`upload-zone ${isDragOver ? 'drag-over' : ''} ${hasFile ? 'has-file' : ''}`}
        onDragOver={(e) => handleDragOver(e, type)}
        onDragLeave={(e) => handleDragLeave(e, type)}
        onDrop={(e) => handleDrop(e, type)}
      >
        <input
          type="file"
          id={type}
          accept={accept}
          onChange={(e) => handleFileChange(type, e)}
          style={{ display: 'none' }}
        />
        <label htmlFor={type} className="upload-label">
          {hasFile ? (
            <>
              <span className="file-icon">&#10003;</span>
              <span className="file-name">{files[type].name}</span>
              <span className="file-size">({(files[type].size / 1024).toFixed(1)} KB)</span>
            </>
          ) : (
            <>
              <span className="upload-icon">&#128194;</span>
              <span className="upload-text">{label}</span>
              <span className="upload-hint">Click or drag to upload</span>
            </>
          )}
        </label>
      </div>
    );
  };

  const allFilesSelected = files.blk && files.rev && files.xor;

  return (
    <div className="upload-container">
      <div className="upload-header">
        <h2>Upload Blockchain Data</h2>
        <p>Upload your blk.dat, rev.dat, and xor.dat files for analysis</p>
      </div>

      <div className="upload-grid">
        {renderDropZone('blk', 'blk.dat', '.dat,*')}
        {renderDropZone('rev', 'rev.dat', '.dat,*')}
        {renderDropZone('xor', 'xor.dat', '.dat,*')}
      </div>

      {error && (
        <div className="upload-error">
          {error}
        </div>
      )}

      <div className="upload-actions">
        <button
          className="upload-button"
          onClick={handleUpload}
          disabled={!allFilesSelected || uploading}
        >
          {uploading ? (
            <>
              <span className="btn-spinner"></span>
              Analyzing...
            </>
          ) : (
            'Analyze Files'
          )}
        </button>
      </div>
    </div>
  );
};
