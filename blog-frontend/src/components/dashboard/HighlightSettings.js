import React, { useEffect, useMemo, useRef, useState } from 'react';
import hljs from '../../highlight/hljs';
import { BASE_KEYS, DEFAULT_PRESET_ID, PRESETS, findPreset } from '../../highlight/palettes';
import { applyHighlight, defaultHighlight } from '../../highlight/theme';
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

const samePalette = (a, b) => BASE_KEYS.every((key) => a.colors[key] === b.colors[key]);

// Lets the admin pick a preset palette, fine-tune individual colors with a
// live preview, and save the result as the site-wide code highlight theme.
function HighlightSettings({ onNotify }) {
  const { highlight, saveHighlight } = useSettings();
  const [draft, setDraft] = useState(highlight);
  const [hexInputs, setHexInputs] = useState(highlight.colors);
  const [saving, setSaving] = useState(false);
  const [showMinor, setShowMinor] = useState(false);
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

  const pickPreset = (preset) => loadPalette({ preset: preset.id, colors: { ...preset.colors } });

  const setColor = (key, value) => {
    setHexInputs((prev) => ({ ...prev, [key]: value }));
    if (HEX_COLOR.test(value)) {
      const colors = { ...draft.colors, [key]: value.toLowerCase() };
      // A palette edited away from its preset is saved as "custom".
      const preset = PRESETS.find((p) => samePalette(p, { colors }));
      setDraft({ preset: preset ? preset.id : 'custom', colors });
    }
  };

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
              {['base0E', 'base0B', 'base09', 'base0D', 'base08', 'base03'].map((key) => (
                <span key={key} style={{ background: preset.colors[key] }} />
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
