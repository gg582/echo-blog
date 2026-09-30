import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import FileManager from '../components/dashboard/FileManager';
import HighlightSettings from '../components/dashboard/HighlightSettings';
import './DashboardPage.css';

const TABS = {
  files: { label: 'Files', Component: FileManager },
  highlight: { label: 'Code colors', Component: HighlightSettings },
};

// How long a success message stays up; errors stay until dismissed.
const NOTICE_MS = 5000;

function DashboardPage() {
  const [params, setParams] = useSearchParams();
  const tab = TABS[params.get('tab')] ? params.get('tab') : 'files';
  const [notice, setNotice] = useState(null);
  const timer = useRef(null);

  const notify = useCallback((message, isError = false) => {
    clearTimeout(timer.current);
    setNotice({ message, isError });
    if (!isError) {
      timer.current = setTimeout(() => setNotice(null), NOTICE_MS);
    }
  }, []);

  useEffect(() => () => clearTimeout(timer.current), []);

  const { Component } = TABS[tab];

  return (
    <div className="dashboard-page">
      <main className="container">
        <h2 className="dashboard-title">Dashboard</h2>

        <nav className="dash-tabs" role="tablist">
          {Object.entries(TABS).map(([key, t]) => (
            <button
              key={key}
              type="button"
              role="tab"
              aria-selected={tab === key}
              className={`dash-tab${tab === key ? ' dash-tab-active' : ''}`}
              onClick={() => { setNotice(null); setParams(key === 'files' ? {} : { tab: key }); }}
            >
              {t.label}
            </button>
          ))}
        </nav>

        {notice && (
          <div className={`dash-alert ${notice.isError ? 'dash-alert-error' : 'dash-alert-ok'}`} role="status">
            <span>{notice.message}</span>
            <button type="button" className="dash-alert-close" onClick={() => setNotice(null)} aria-label="Dismiss">×</button>
          </div>
        )}

        <Component onNotify={notify} />
      </main>
    </div>
  );
}

export default DashboardPage;
