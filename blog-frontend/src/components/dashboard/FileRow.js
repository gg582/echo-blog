import React, { useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { assetUrl } from '../../api/admin';
import { copyText, extension, fileKind, formatDate, formatSize, markdownFor, validFileName } from './fileUtils';

// One file in the manager: preview, name, usage, metadata and actions.
// Rename happens inline; the other actions are delegated to the parent.
function FileRow({ file, usedBy, selected, busy, onSelect, onDelete, onReplace, onRename, onNotify }) {
  const replaceInput = useRef(null);
  const [renaming, setRenaming] = useState(false);
  const [newName, setNewName] = useState(file.name);
  const [updateRefs, setUpdateRefs] = useState(true);

  const url = assetUrl(file.name);
  const previewUrl = assetUrl(file.name, file.modifiedAt);
  const kind = fileKind(file.name);

  const copy = async (text, what) => {
    try {
      await copyText(text);
      onNotify(`${what} copied.`);
    } catch (e) {
      onNotify(`Could not copy: ${e.message}`, true);
    }
  };

  const startRename = () => {
    setNewName(file.name);
    setUpdateRefs(true);
    setRenaming(true);
  };

  const submitRename = async (event) => {
    event.preventDefault();
    const target = newName.trim();
    if (target === file.name) {
      setRenaming(false);
      return;
    }
    if (!validFileName(target)) {
      onNotify('File names cannot be empty or contain "/", "\\" or "..".', true);
      return;
    }
    if (await onRename(file, target, usedBy.length > 0 && updateRefs)) {
      setRenaming(false);
    }
  };

  const chooseReplacement = (event) => {
    const replacement = event.target.files && event.target.files[0];
    event.target.value = '';
    if (!replacement) { return; }
    if (extension(replacement.name) !== extension(file.name) &&
        !window.confirm(`"${replacement.name}" has a different file type than "${file.name}". Replace anyway? The file keeps the name "${file.name}".`)) {
      return;
    }
    onReplace(file, replacement);
  };

  return (
    <tr className={selected ? 'fm-row-selected' : undefined}>
      <td className="fm-check">
        <input type="checkbox" checked={selected} onChange={(e) => onSelect(file.name, e.target.checked)} aria-label={`Select ${file.name}`} />
      </td>
      <td className="fm-preview">
        <a href={url} target="_blank" rel="noreferrer">
          {kind === 'image'
            ? <img src={previewUrl} alt="" loading="lazy" />
            : <span className={`fm-kind fm-kind-${kind}`}>{extension(file.name) || kind}</span>}
        </a>
      </td>
      <td className="fm-name">
        {renaming ? (
          <form className="fm-rename" onSubmit={submitRename}>
            <input
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => { if (e.key === 'Escape') { setRenaming(false); } }}
              autoFocus
              disabled={busy}
            />
            {usedBy.length > 0 && (
              <label className="fm-rename-refs">
                <input type="checkbox" checked={updateRefs} onChange={(e) => setUpdateRefs(e.target.checked)} />
                Update links in {usedBy.length} post{usedBy.length > 1 ? 's' : ''}
              </label>
            )}
            <div className="fm-rename-actions">
              <button type="submit" className="btn btn-primary" disabled={busy}>Save</button>
              <button type="button" className="btn" onClick={() => setRenaming(false)} disabled={busy}>Cancel</button>
            </div>
          </form>
        ) : (
          <>
            <a href={url} target="_blank" rel="noreferrer" className="fm-filename">{file.name}</a>
            <div className="fm-usage">
              {usedBy.length === 0
                ? <span className="fm-badge fm-badge-unused">not used</span>
                : usedBy.map((id) => (
                    <Link key={id} to={`/posts/${encodeURIComponent(id)}`} className="fm-badge" title={`Used in ${id}`}>{id}</Link>
                  ))}
            </div>
          </>
        )}
      </td>
      <td className="fm-meta">
        <div>{formatSize(file.size)}</div>
        <div className="fm-date">{formatDate(file.modifiedAt)}</div>
      </td>
      <td className="fm-actions">
        <button type="button" className="btn" onClick={() => copy(url, 'Link')} disabled={busy}>Copy link</button>
        <button type="button" className="btn" onClick={() => copy(markdownFor(file.name, url), 'Markdown')} disabled={busy}>Copy Markdown</button>
        <button type="button" className="btn" onClick={() => replaceInput.current.click()} disabled={busy}>Replace</button>
        <input ref={replaceInput} type="file" hidden onChange={chooseReplacement} />
        <button type="button" className="btn" onClick={startRename} disabled={busy || renaming}>Rename</button>
        <button type="button" className="btn btn-danger" onClick={() => onDelete([file.name])} disabled={busy}>Delete</button>
      </td>
    </tr>
  );
}

export default FileRow;
