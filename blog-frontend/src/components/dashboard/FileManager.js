import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { deleteFile, fileUsage, listFiles, renameFile, replaceFile, uploadFiles } from '../../api/admin';
import FileRow from './FileRow';
import UploadDropzone from './UploadDropzone';
import { FILTERS, SORTS, formatSize } from './fileUtils';

// Lists uploaded files with search, filters and bulk selection, and handles
// upload, replace, rename and delete.
function FileManager({ onNotify }) {
  const [files, setFiles] = useState([]);
  const [usage, setUsage] = useState({});
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(null);
  const [busy, setBusy] = useState(false);
  const [query, setQuery] = useState('');
  const [sort, setSort] = useState('newest');
  const [filter, setFilter] = useState('all');
  const [selected, setSelected] = useState(() => new Set());

  const refresh = useCallback(async () => {
    try {
      const [list, used] = await Promise.all([listFiles(), fileUsage()]);
      setFiles(list || []);
      setUsage(used || {});
      setLoadError(null);
      // Drop selections of files that no longer exist.
      setSelected((prev) => new Set([...prev].filter((name) => (list || []).some((f) => f.name === name))));
    } catch (e) {
      setLoadError(e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  // run executes an action with the manager locked, reports the outcome and
  // reloads the list. It returns whether the action succeeded.
  const run = async (action, successMessage) => {
    setBusy(true);
    try {
      const result = await action();
      const message = typeof successMessage === 'function' ? successMessage(result) : successMessage;
      if (message) { onNotify(message); }
      return true;
    } catch (e) {
      onNotify(e.message, true);
      return false;
    } finally {
      await refresh();
      setBusy(false);
    }
  };

  const visible = useMemo(() => {
    const q = query.trim().toLowerCase();
    return files
      .filter((f) => !q || f.name.toLowerCase().includes(q))
      .filter((f) => FILTERS[filter].match(f, usage))
      .sort(SORTS[sort].compare);
  }, [files, usage, query, filter, sort]);

  const totalSize = files.reduce((sum, f) => sum + f.size, 0);
  const unusedCount = files.filter((f) => !(usage[f.name] || []).length).length;
  const allVisibleSelected = visible.length > 0 && visible.every((f) => selected.has(f.name));

  const select = (name, on) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (on) { next.add(name); } else { next.delete(name); }
      return next;
    });
  };

  const selectAllVisible = (on) => {
    setSelected((prev) => {
      const next = new Set(prev);
      visible.forEach((f) => (on ? next.add(f.name) : next.delete(f.name)));
      return next;
    });
  };

  const handleUpload = (chosen) => {
    const existing = chosen.filter((f) => files.some((x) => x.name === f.name)).map((f) => f.name);
    if (existing.length > 0 &&
        !window.confirm(`These files already exist and will be overwritten:\n\n${existing.join('\n')}\n\nContinue?`)) {
      return;
    }
    run(() => uploadFiles(chosen), ({ partial, data }) =>
      partial
        ? `Uploaded ${data.length} of ${chosen.length} files; the rest failed (see server log).`
        : `Uploaded ${data.length} file${data.length === 1 ? '' : 's'}.`);
  };

  const handleDelete = (names) => {
    const used = names.filter((name) => (usage[name] || []).length);
    let message = names.length === 1 ? `Delete "${names[0]}"?` : `Delete ${names.length} files?`;
    if (used.length > 0) {
      const detail = used.map((name) => `${name} → ${usage[name].join(', ')}`).join('\n');
      message += `\n\nThese are still linked from posts, which will show broken links:\n\n${detail}`;
    }
    if (!window.confirm(message)) { return; }

    run(async () => {
      const failed = [];
      for (const name of names) {
        try {
          await deleteFile(name);
        } catch (e) {
          failed.push(`${name}: ${e.message}`);
        }
      }
      if (failed.length > 0) {
        throw new Error(`Some files could not be deleted:\n${failed.join('\n')}`);
      }
    }, `Deleted ${names.length} file${names.length === 1 ? '' : 's'}.`);
  };

  const handleReplace = (file, replacement) =>
    run(() => replaceFile(file.name, replacement),
      `Replaced the content of "${file.name}" (${formatSize(replacement.size)}).`);

  const handleRename = (file, to, updateReferences) =>
    run(() => renameFile(file.name, to, updateReferences), (result) => {
      const posts = result.updatedPosts || [];
      return posts.length > 0
        ? `Renamed to "${to}" and updated links in: ${posts.join(', ')}.`
        : `Renamed to "${to}".`;
    });

  if (loading) {
    return <div className="loading-spinner">Loading files...</div>;
  }

  const selectedNames = [...selected];

  return (
    <section className="fm">
      <UploadDropzone onFiles={handleUpload} busy={busy} />

      {loadError && <div className="dash-alert dash-alert-error">{loadError.message}</div>}

      <div className="fm-toolbar">
        <input
          type="search"
          className="fm-search"
          placeholder="Search files…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <select value={filter} onChange={(e) => setFilter(e.target.value)} aria-label="Filter">
          {Object.entries(FILTERS).map(([key, f]) => <option key={key} value={key}>{f.label}</option>)}
        </select>
        <select value={sort} onChange={(e) => setSort(e.target.value)} aria-label="Sort">
          {Object.entries(SORTS).map(([key, s]) => <option key={key} value={key}>{s.label}</option>)}
        </select>
        <button type="button" className="btn" onClick={refresh} disabled={busy}>Refresh</button>
      </div>

      <div className="fm-summary">
        <span>
          {files.length} files · {formatSize(totalSize)} · {unusedCount} not used in any post
          {visible.length !== files.length && ` · showing ${visible.length}`}
        </span>
        {selectedNames.length > 0 && (
          <span className="fm-bulk">
            {selectedNames.length} selected
            <button type="button" className="btn btn-danger" onClick={() => handleDelete(selectedNames)} disabled={busy}>
              Delete selected
            </button>
            <button type="button" className="btn" onClick={() => setSelected(new Set())} disabled={busy}>Clear</button>
          </span>
        )}
      </div>

      {visible.length === 0 ? (
        <div className="empty-state">
          <p>{files.length === 0 ? 'No uploaded files found.' : 'No files match the current search or filter.'}</p>
        </div>
      ) : (
        <div className="fm-table-wrap">
          <table className="fm-table">
            <thead>
              <tr>
                <th className="fm-check">
                  <input type="checkbox" checked={allVisibleSelected} onChange={(e) => selectAllVisible(e.target.checked)} aria-label="Select all shown files" />
                </th>
                <th className="fm-preview"></th>
                <th>Name / used in</th>
                <th className="fm-meta">Size / modified</th>
                <th className="fm-actions"></th>
              </tr>
            </thead>
            <tbody>
              {visible.map((file) => (
                <FileRow
                  key={file.name}
                  file={file}
                  usedBy={usage[file.name] || []}
                  selected={selected.has(file.name)}
                  busy={busy}
                  onSelect={select}
                  onDelete={handleDelete}
                  onReplace={handleReplace}
                  onRename={handleRename}
                  onNotify={onNotify}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

export default FileManager;
