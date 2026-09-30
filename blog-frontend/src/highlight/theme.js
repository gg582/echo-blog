import { BASE_KEYS, DEFAULT_PRESET_ID, findPreset } from './palettes';
import './theme.css';

const HEX_COLOR = /^#[0-9a-f]{6}$/i;

// The highlight theme used when the server has no saved setting.
export const defaultHighlight = () => {
  const preset = findPreset(DEFAULT_PRESET_ID);
  return { preset: preset.id, colors: { ...preset.colors } };
};

// Fill in missing or malformed colors from the default palette.
export const normalizeHighlight = (highlight) => {
  const base = defaultHighlight();
  if (!highlight || typeof highlight !== 'object') {
    return base;
  }
  const colors = { ...base.colors };
  for (const key of BASE_KEYS) {
    const value = highlight.colors && highlight.colors[key];
    if (typeof value === 'string' && HEX_COLOR.test(value)) {
      colors[key] = value.toLowerCase();
    }
  }
  return { preset: typeof highlight.preset === 'string' ? highlight.preset : base.preset, colors };
};

// Set the --hl-baseXX variables on element (the document root by default),
// which restyles every highlighted code block on the page.
export const applyHighlight = (highlight, element = document.documentElement) => {
  const { colors } = normalizeHighlight(highlight);
  for (const key of BASE_KEYS) {
    element.style.setProperty(`--hl-${key}`, colors[key]);
  }
};
