import React, { useRef, useState } from 'react';

// A drop area that also opens the file picker when clicked. Calls onFiles
// with the chosen File objects.
function UploadDropzone({ onFiles, busy }) {
  const inputRef = useRef(null);
  const [dragging, setDragging] = useState(false);

  const handleFiles = (fileList) => {
    const files = Array.from(fileList || []);
    if (files.length > 0) {
      onFiles(files);
    }
  };

  const onDrop = (event) => {
    event.preventDefault();
    setDragging(false);
    if (!busy) {
      handleFiles(event.dataTransfer.files);
    }
  };

  return (
    <div
      className={`dz${dragging ? ' dz-active' : ''}${busy ? ' dz-busy' : ''}`}
      role="button"
      tabIndex={0}
      onClick={() => !busy && inputRef.current.click()}
      onKeyDown={(e) => { if ((e.key === 'Enter' || e.key === ' ') && !busy) { e.preventDefault(); inputRef.current.click(); } }}
      onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
      onDragLeave={() => setDragging(false)}
      onDrop={onDrop}
    >
      <input
        ref={inputRef}
        type="file"
        multiple
        hidden
        onChange={(e) => { handleFiles(e.target.files); e.target.value = ''; }}
      />
      {busy ? 'Uploading…' : <>Drop files here or <span className="dz-link">browse</span> to upload</>}
    </div>
  );
}

export default UploadDropzone;
