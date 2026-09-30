// src/context/SettingsContext.js
// Loads the site settings once and applies them (currently the code highlight
// palette) to every page. The dashboard saves new settings through here so
// the change shows up immediately.
import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import API_BASE_URL from '../config/api';
import { saveSettings } from '../api/admin';
import { applyHighlight, defaultHighlight, normalizeHighlight } from '../highlight/theme';

const SettingsContext = createContext(null);

export const SettingsProvider = ({ children }) => {
  const [settings, setSettings] = useState({ highlight: null });
  const highlight = useMemo(
    () => (settings.highlight ? normalizeHighlight(settings.highlight) : defaultHighlight()),
    [settings.highlight],
  );

  useEffect(() => {
    let cancelled = false;
    fetch(`${API_BASE_URL}/api/settings`)
      .then((response) => (response.ok ? response.json() : null))
      .then((loaded) => {
        if (!cancelled && loaded) {
          setSettings(loaded);
        }
      })
      // Without settings the default palette from theme.css stays in place.
      .catch(() => {});
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    applyHighlight(highlight);
  }, [highlight]);

  const saveHighlight = useCallback(async (next) => {
    const saved = await saveSettings({ ...settings, highlight: normalizeHighlight(next) });
    setSettings(saved);
  }, [settings]);

  return (
    <SettingsContext.Provider value={{ highlight, saveHighlight }}>
      {children}
    </SettingsContext.Provider>
  );
};

export const useSettings = () => {
  const context = useContext(SettingsContext);
  if (context === null) {
    throw new Error('useSettings must be used within a SettingsProvider');
  }
  return context;
};
