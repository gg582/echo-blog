import { BASE_KEYS, DEFAULT_PRESET_ID, findPreset } from './palettes';
import './theme.css';

const HEX_COLOR = /^#[0-9a-f]{6}$/i;

// Tokens a palette can style individually (see the end of theme.css), with
// the base16 slot each falls back to.
export const TOKENS = {
  comment: { label: 'Comments', slot: 'base03' },
  keyword: { label: 'Keywords', slot: 'base0E' },
  type: { label: 'Types', slot: 'base0E' },
  built_in: { label: 'Built-ins', slot: 'base0C' },
  function: { label: 'Function names', slot: 'base0D' },
  title: { label: 'Class names', slot: 'base0A' },
  string: { label: 'Strings', slot: 'base0B' },
  number: { label: 'Numbers', slot: 'base09' },
  literal: { label: 'Literals (true, nil)', slot: 'base09' },
  variable: { label: 'Variables', slot: 'base08' },
  attr: { label: 'Attributes, keys', slot: 'base09' },
  meta: { label: 'Preprocessor, meta', slot: 'base0F' },
  section: { label: 'Sections, headings', slot: 'base0D' },
  operator: { label: 'Operators', slot: 'base05' },
  punctuation: { label: 'Punctuation', slot: 'base05' },
};

// The highlight theme used when the server has no saved setting.
export const defaultHighlight = () => {
  const preset = findPreset(DEFAULT_PRESET_ID);
  return { preset: preset.id, colors: { ...preset.colors }, tokens: {} };
};

// Keep only well-formed token overrides: { color?: '#rrggbb', bold?, italic? }.
const normalizeTokens = (tokens) => {
  const out = {};
  if (!tokens || typeof tokens !== 'object') {
    return out;
  }
  for (const name of Object.keys(TOKENS)) {
    const t = tokens[name];
    if (!t || typeof t !== 'object') { continue; }
    const style = {};
    if (typeof t.color === 'string' && HEX_COLOR.test(t.color)) { style.color = t.color.toLowerCase(); }
    if (t.bold === true) { style.bold = true; }
    if (t.italic === true) { style.italic = true; }
    if (Object.keys(style).length > 0) { out[name] = style; }
  }
  return out;
};

// Fill in missing or malformed colors from the default palette and drop
// malformed token overrides.
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
  return {
    preset: typeof highlight.preset === 'string' ? highlight.preset : base.preset,
    colors,
    tokens: normalizeTokens(highlight.tokens),
  };
};

// Whether two palettes render identically.
export const samePalette = (a, b) =>
  BASE_KEYS.every((key) => a.colors[key] === b.colors[key]) &&
  JSON.stringify(normalizeTokens(a.tokens)) === JSON.stringify(normalizeTokens(b.tokens));

// Set the --hl-* variables on element (the document root by default), which
// restyles every highlighted code block inside it. Token variables that the
// palette does not override are removed so the base16 fallbacks apply.
export const applyHighlight = (highlight, element = document.documentElement) => {
  const { colors, tokens } = normalizeHighlight(highlight);
  for (const key of BASE_KEYS) {
    element.style.setProperty(`--hl-${key}`, colors[key]);
  }
  for (const name of Object.keys(TOKENS)) {
    const t = tokens[name] || {};
    const set = (prop, value) => (value
      ? element.style.setProperty(prop, value)
      : element.style.removeProperty(prop));
    set(`--hl-${name}`, t.color);
    set(`--hl-${name}-weight`, t.bold ? 'bold' : null);
    set(`--hl-${name}-style`, t.italic ? 'italic' : null);
  }
};
