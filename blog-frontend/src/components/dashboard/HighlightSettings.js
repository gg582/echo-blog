import React, { useEffect, useMemo, useRef, useState } from 'react';
import hljs from '../../highlight/hljs';
import { BASE_KEYS, DEFAULT_PRESET_ID, PRESETS, findPreset } from '../../highlight/palettes';
import { TOKENS, applyHighlight, defaultHighlight, samePalette } from '../../highlight/theme';
import { useSettings } from '../../context/SettingsContext';

// What each base16 slot colors, following the base16 styling guidelines.
const SLOT_ROLES = {
  base00: 'Background',
  base01: 'Lighter background',
  base02: 'Selection',
  base03: 'Comments',
  base04: 'Markup tags',
  base05: 'Text, operators',
  base06: 'Light foreground',
  base07: 'Light background',
  base08: 'Variables, XML tags',
  base09: 'Numbers, constants, attributes',
  base0A: 'Classes, bold',
  base0B: 'Strings',
  base0C: 'Built-ins, regex, escapes',
  base0D: 'Functions, headings',
  base0E: 'Keywords, types',
  base0F: 'Meta, embedded tags',
};

// Slots that are rarely visible in highlight.js output; shown collapsed.
const MINOR_SLOTS = ['base01', 'base06', 'base07'];

const SAMPLES = [
  {
    language: 'go',
    code: `// Serve starts the blog server.
func Serve(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, ReadTimeout: 5 * time.Second}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve %s: %w", addr, err)
	}
	return nil
}`,
  },
  {
    language: 'javascript',
    code: `class PostList extends Component {
  async load(page = 1) {
    const res = await fetch(\`/api/posts?page=\${page}\`, { method: 'POST' });
    return /^2\\d\\d$/.test(String(res.status)) ? res.json() : null; // 2xx only
  }
}`,
  },
  {
    language: 'desktop',
    code: `[Desktop Entry]
# Launcher for the blog editor
Name=Echo Blog
Exec=firefox --new-window https://chatter.pw/new-post
Terminal=false
Categories=Network;WebBrowser;`,
  },
];

const HEX_COLOR = /^#[0-9a-f]{6}$/i;

// The color a token actually renders in: its override or its base16 slot.
const tokenColor = (palette, name) =>
  (palette.tokens && palette.tokens[name] && palette.tokens[name].color) || palette.colors[TOKENS[name].slot];

// Tokens shown in the preset swatches.
const SWATCH_TOKENS = ['keyword', 'type', 'string', 'number', 'function', 'comment'];

const copyPalette = (p) => ({
  preset: p.id || p.preset,
  colors: { ...p.colors },
  tokens: JSON.parse(JSON.stringify(p.tokens || {})),
});

// The preset a palette matches exactly, or "custom".
const presetIdOf = (palette) => {
  const match = PRESETS.find((p) => samePalette(p, palette));
  return match ? match.id : 'custom';
};

// Lets the admin pick a preset palette, fine-tune individual colors with a
// live preview, and save the result as the site-wide code highlight theme.
function HighlightSettings({ onNotify }) {
  const { highlight, saveHighlight } = useSettings();
  const [draft, setDraft] = useState(highlight);
  const [hexInputs, setHexInputs] = useState(highlight.colors);
  const [saving, setSaving] = useState(false);
  const [showMinor, setShowMinor] = useState(false);
  const [showTokens, setShowTokens] = useState(false);
  const previewRef = useRef(null);

  // Follow the saved theme when it loads or changes elsewhere.
  useEffect(() => {
    setDraft(highlight);
    setHexInputs(highlight.colors);
  }, [highlight]);

  // The preview uses its own copy of the variables, so the rest of the site
  // keeps the saved theme until the draft is saved.
  useEffect(() => {
    if (previewRef.current) {
      applyHighlight(draft, previewRef.current);
    }
  }, [draft]);

  const samples = useMemo(() => SAMPLES.map((s) => ({
    ...s,
    html: hljs.highlight(s.code, { language: s.language }).value,
  })), []);

  const dirty = !samePalette(draft, highlight);
  const draftPreset = findPreset(draft.preset);
  const isDefault = draft.preset === DEFAULT_PRESET_ID && samePalette(draft, defaultHighlight());

  const loadPalette = (next) => {
    setDraft(next);
    setHexInputs(next.colors);
  };

  const pickPreset = (preset) => loadPalette(copyPalette(preset));

  // updateDraft applies an edit; a palette edited away from its preset is
  // saved as "custom".
  const updateDraft = (colors, tokens) => {
    const next = { colors, tokens };
    setDraft({ ...next, preset: presetIdOf(next) });
  };

  const setColor = (key, value) => {
    setHexInputs((prev) => ({ ...prev, [key]: value }));
    if (HEX_COLOR.test(value)) {
      updateDraft({ ...draft.colors, [key]: value.toLowerCase() }, draft.tokens);
    }
  };

  // setToken merges a change into one token's override; an override with
  // nothing left is removed so the token follows its base16 slot again.
  const setToken = (name, change) => {
    const merged = { ...(draft.tokens[name] || {}), ...change };
    Object.keys(merged).forEach((k) => { if (merged[k] === undefined || merged[k] === false) { delete merged[k]; } });
    const tokens = { ...draft.tokens };
    if (Object.keys(merged).length > 0) { tokens[name] = merged; } else { delete tokens[name]; }
    updateDraft(draft.colors, tokens);
  };

  const overriddenCount = Object.keys(draft.tokens || {}).length;

  const save = async () => {
    setSaving(true);
    try {
      await saveHighlight(draft);
      onNotify('Code highlight colors saved. Every page now uses them.');
    } catch (e) {
      onNotify(`Could not save: ${e.message}`, true);
    } finally {
      setSaving(false);
    }
  };

  const slots = BASE_KEYS.filter((key) => showMinor || !MINOR_SLOTS.includes(key));

  return (
    <section className="hl">
      <div className="hl-presets">
        {PRESETS.map((preset) => (
          <button
            type="button"
            key={preset.id}
            className={`hl-preset${draft.preset === preset.id ? ' hl-preset-active' : ''}`}
            onClick={() => pickPreset(preset)}
            title={preset.label}
          >
            <span className="hl-swatch" style={{ background: preset.colors.base00 }}>
              {SWATCH_TOKENS.map((name) => (
                <span key={name} style={{ background: tokenColor(preset, name) }} />
              ))}
            </span>
            <span className="hl-preset-label">{preset.label}</span>
          </button>
        ))}
      </div>

      <div className="hl-editor">
        <div className="hl-colors">
          <h3>
            Colors
            <span className="hl-current">
              {draftPreset ? draftPreset.label : 'Custom palette'}
            </span>
          </h3>
          {slots.map((key) => (
            <label key={key} className="hl-color">
              <input
                type="color"
                value={draft.colors[key]}
                onChange={(e) => setColor(key, e.target.value)}
                aria-label={`${SLOT_ROLES[key]} color`}
              />
              <input
                type="text"
                className={`hl-hex${HEX_COLOR.test(hexInputs[key]) ? '' : ' hl-hex-invalid'}`}
                value={hexInputs[key]}
                onChange={(e) => setColor(key, e.target.value.trim())}
                spellCheck={false}
                maxLength={7}
              />
              <span className="hl-role">{SLOT_ROLES[key]}</span>
            </label>
          ))}
          <button type="button" className="btn btn-link" onClick={() => setShowMinor(!showMinor)}>
            {showMinor ? 'Hide rarely used colors' : 'Show all 16 colors'}
          </button>

          <h3 className="hl-tokens-title">
            Token styles
            <span className="hl-current">{overriddenCount} overridden</span>
          </h3>
          <p className="hl-hint">
            Style single tokens beyond the 16 colors, e.g. types apart from keywords, or bold keywords.
            Unchecked colors follow the slot shown.
          </p>
          {showTokens && Object.entries(TOKENS).map(([name, token]) => {
            const t = draft.tokens[name] || {};
            return (
              <div key={name} className="hl-token">
                <input
                  type="checkbox"
                  checked={!!t.color}
                  onChange={(e) => setToken(name, { color: e.target.checked ? tokenColor(draft, name) : undefined })}
                  aria-label={`Own color for ${token.label}`}
                />
                <input
                  type="color"
                  value={tokenColor(draft, name)}
                  disabled={!t.color}
                  onChange={(e) => setToken(name, { color: e.target.value })}
                  aria-label={`${token.label} color`}
                />
                <button
                  type="button"
                  className={`hl-style${t.bold ? ' hl-style-on' : ''}`}
                  onClick={() => setToken(name, { bold: !t.bold })}
                  aria-pressed={!!t.bold}
                  title="Bold"
                ><b>B</b></button>
                <button
                  type="button"
                  className={`hl-style${t.italic ? ' hl-style-on' : ''}`}
                  onClick={() => setToken(name, { italic: !t.italic })}
                  aria-pressed={!!t.italic}
                  title="Italic"
                ><i>I</i></button>
                <span className="hl-role">
                  {token.label}
                  {!t.color && <span className="hl-slot"> · {token.slot}</span>}
                </span>
              </div>
            );
          })}
          <button type="button" className="btn btn-link" onClick={() => setShowTokens(!showTokens)}>
            {showTokens ? 'Hide token styles' : 'Edit token styles'}
          </button>
        </div>

        <div className="hl-preview" ref={previewRef}>
          <h3>Preview</h3>
          {samples.map((s) => (
            <pre key={s.language}>
              <code className={`hljs language-${s.language}`} dangerouslySetInnerHTML={{ __html: s.html }} />
            </pre>
          ))}
        </div>
      </div>

      <div className="hl-actions">
        <button type="button" className="btn btn-primary" onClick={save} disabled={!dirty || saving}>
          {saving ? 'Saving…' : 'Save for the whole site'}
        </button>
        <button type="button" className="btn" onClick={() => loadPalette(highlight)} disabled={!dirty || saving}>
          Discard changes
        </button>
        <button type="button" className="btn" onClick={() => loadPalette(defaultHighlight())} disabled={isDefault || saving}>
          Reset to default
        </button>
        {dirty && <span className="hl-unsaved">Unsaved changes. Only this preview shows them.</span>}
      </div>
    </section>
  );
}

export default HighlightSettings;
