// Base16 palettes for code highlighting, extracted from the scheme headers of
// highlight.js/styles/base16/*.css (highlight.js 11.12.0).
// Each palette maps base00..base0F to a color; src/highlight/theme.css maps
// highlight.js token classes to these slots the same way the base16 themes do.
// A palette may also have `tokens`: per-token overrides (see TOKENS in theme.js).

export const BASE_KEYS = [
  'base00', 'base01', 'base02', 'base03', 'base04', 'base05', 'base06', 'base07',
  'base08', 'base09', 'base0A', 'base0B', 'base0C', 'base0D', 'base0E', 'base0F',
];

const BASE16_PRESETS = [
  {
    id: 'github',
    label: 'GitHub (default)',
    dark: false,
    colors: {
      // base16 github uses #ffffff; code blocks have always shown #f6f8fa here.
      base00: '#f6f8fa',
      base01: '#f5f5f5',
      base02: '#c8c8fa',
      base03: '#969896',
      base04: '#e8e8e8',
      base05: '#333333',
      base06: '#ffffff',
      base07: '#ffffff',
      base08: '#ed6a43',
      base09: '#0086b3',
      base0A: '#795da3',
      base0B: '#183691',
      base0C: '#183691',
      base0D: '#795da3',
      base0E: '#a71d5d',
      base0F: '#333333',
    },
  },
  {
    id: 'one-light',
    label: 'One Light',
    dark: false,
    colors: {
      base00: '#fafafa',
      base01: '#f0f0f1',
      base02: '#e5e5e6',
      base03: '#a0a1a7',
      base04: '#696c77',
      base05: '#383a42',
      base06: '#202227',
      base07: '#090a0b',
      base08: '#ca1243',
      base09: '#d75f00',
      base0A: '#c18401',
      base0B: '#50a14f',
      base0C: '#0184bc',
      base0D: '#4078f2',
      base0E: '#a626a4',
      base0F: '#986801',
    },
  },
  {
    id: 'solarized-light',
    label: 'Solarized Light',
    dark: false,
    colors: {
      base00: '#fdf6e3',
      base01: '#eee8d5',
      base02: '#93a1a1',
      base03: '#839496',
      base04: '#657b83',
      base05: '#586e75',
      base06: '#073642',
      base07: '#002b36',
      base08: '#dc322f',
      base09: '#cb4b16',
      base0A: '#b58900',
      base0B: '#859900',
      base0C: '#2aa198',
      base0D: '#268bd2',
      base0E: '#6c71c4',
      base0F: '#d33682',
    },
  },
  {
    id: 'gruvbox-light-medium',
    label: 'Gruvbox Light',
    dark: false,
    colors: {
      base00: '#fbf1c7',
      base01: '#ebdbb2',
      base02: '#d5c4a1',
      base03: '#bdae93',
      base04: '#665c54',
      base05: '#504945',
      base06: '#3c3836',
      base07: '#282828',
      base08: '#9d0006',
      base09: '#af3a03',
      base0A: '#b57614',
      base0B: '#79740e',
      base0C: '#427b58',
      base0D: '#076678',
      base0E: '#8f3f71',
      base0F: '#d65d0e',
    },
  },
  {
    id: 'material-lighter',
    label: 'Material Lighter',
    dark: false,
    colors: {
      base00: '#fafafa',
      base01: '#e7eaec',
      base02: '#cceae7',
      base03: '#ccd7da',
      base04: '#8796b0',
      base05: '#80cbc4',
      base06: '#80cbc4',
      base07: '#ffffff',
      base08: '#ff5370',
      base09: '#f76d47',
      base0A: '#ffb62c',
      base0B: '#91b859',
      base0C: '#39adb5',
      base0D: '#6182b8',
      base0E: '#7c4dff',
      base0F: '#e53935',
    },
  },
  {
    id: 'tomorrow',
    label: 'Tomorrow',
    dark: false,
    colors: {
      base00: '#ffffff',
      base01: '#e0e0e0',
      base02: '#d6d6d6',
      base03: '#8e908c',
      base04: '#969896',
      base05: '#4d4d4c',
      base06: '#282a2e',
      base07: '#1d1f21',
      base08: '#c82829',
      base09: '#f5871f',
      base0A: '#eab700',
      base0B: '#718c00',
      base0C: '#3e999f',
      base0D: '#4271ae',
      base0E: '#8959a8',
      base0F: '#a3685a',
    },
  },
  {
    id: 'atelier-forest-light',
    label: 'Atelier Forest Light',
    dark: false,
    colors: {
      base00: '#f1efee',
      base01: '#e6e2e0',
      base02: '#a8a19f',
      base03: '#9c9491',
      base04: '#766e6b',
      base05: '#68615e',
      base06: '#2c2421',
      base07: '#1b1918',
      base08: '#f22c40',
      base09: '#df5320',
      base0A: '#c38418',
      base0B: '#7b9726',
      base0C: '#3d97b8',
      base0D: '#407ee7',
      base0E: '#6666ea',
      base0F: '#c33ff3',
    },
  },
  {
    id: 'monokai',
    label: 'Monokai',
    dark: true,
    colors: {
      base00: '#272822',
      base01: '#383830',
      base02: '#49483e',
      base03: '#75715e',
      base04: '#a59f85',
      base05: '#f8f8f2',
      base06: '#f5f4f1',
      base07: '#f9f8f5',
      base08: '#f92672',
      base09: '#fd971f',
      base0A: '#f4bf75',
      base0B: '#a6e22e',
      base0C: '#a1efe4',
      base0D: '#66d9ef',
      base0E: '#ae81ff',
      base0F: '#cc6633',
    },
  },
  {
    id: 'dracula',
    label: 'Dracula',
    dark: true,
    colors: {
      base00: '#282936',
      base01: '#3a3c4e',
      base02: '#4d4f68',
      base03: '#626483',
      base04: '#62d6e8',
      base05: '#e9e9f4',
      base06: '#f1f2f8',
      base07: '#f7f7fb',
      base08: '#ea51b2',
      base09: '#b45bcf',
      base0A: '#00f769',
      base0B: '#ebff87',
      base0C: '#a1efe4',
      base0D: '#62d6e8',
      base0E: '#b45bcf',
      base0F: '#00f769',
    },
  },
  {
    id: 'nord',
    label: 'Nord',
    dark: true,
    colors: {
      base00: '#2e3440',
      base01: '#3b4252',
      base02: '#434c5e',
      base03: '#4c566a',
      base04: '#d8dee9',
      base05: '#e5e9f0',
      base06: '#eceff4',
      base07: '#8fbcbb',
      base08: '#bf616a',
      base09: '#d08770',
      base0A: '#ebcb8b',
      base0B: '#a3be8c',
      base0C: '#88c0d0',
      base0D: '#81a1c1',
      base0E: '#b48ead',
      base0F: '#5e81ac',
    },
  },
  {
    id: 'onedark',
    label: 'One Dark',
    dark: true,
    colors: {
      base00: '#282c34',
      base01: '#353b45',
      base02: '#3e4451',
      base03: '#545862',
      base04: '#565c64',
      base05: '#abb2bf',
      base06: '#b6bdca',
      base07: '#c8ccd4',
      base08: '#e06c75',
      base09: '#d19a66',
      base0A: '#e5c07b',
      base0B: '#98c379',
      base0C: '#56b6c2',
      base0D: '#61afef',
      base0E: '#c678dd',
      base0F: '#be5046',
    },
  },
  {
    id: 'solarized-dark',
    label: 'Solarized Dark',
    dark: true,
    colors: {
      base00: '#002b36',
      base01: '#073642',
      base02: '#586e75',
      base03: '#657b83',
      base04: '#839496',
      base05: '#93a1a1',
      base06: '#eee8d5',
      base07: '#fdf6e3',
      base08: '#dc322f',
      base09: '#cb4b16',
      base0A: '#b58900',
      base0B: '#859900',
      base0C: '#2aa198',
      base0D: '#268bd2',
      base0E: '#6c71c4',
      base0F: '#d33682',
    },
  },
  {
    id: 'gruvbox-dark-medium',
    label: 'Gruvbox Dark',
    dark: true,
    colors: {
      base00: '#282828',
      base01: '#3c3836',
      base02: '#504945',
      base03: '#665c54',
      base04: '#bdae93',
      base05: '#d5c4a1',
      base06: '#ebdbb2',
      base07: '#fbf1c7',
      base08: '#fb4934',
      base09: '#fe8019',
      base0A: '#fabd2f',
      base0B: '#b8bb26',
      base0C: '#8ec07c',
      base0D: '#83a598',
      base0E: '#d3869b',
      base0F: '#d65d0e',
    },
  },
  {
    id: 'material-darker',
    label: 'Material Darker',
    dark: true,
    colors: {
      base00: '#212121',
      base01: '#303030',
      base02: '#353535',
      base03: '#4a4a4a',
      base04: '#b2ccd6',
      base05: '#eeffff',
      base06: '#eeffff',
      base07: '#ffffff',
      base08: '#f07178',
      base09: '#f78c6c',
      base0A: '#ffcb6b',
      base0B: '#c3e88d',
      base0C: '#89ddff',
      base0D: '#82aaff',
      base0E: '#c792ea',
      base0F: '#ff5370',
    },
  },
  {
    id: 'ocean',
    label: 'Ocean',
    dark: true,
    colors: {
      base00: '#2b303b',
      base01: '#343d46',
      base02: '#4f5b66',
      base03: '#65737e',
      base04: '#a7adba',
      base05: '#c0c5ce',
      base06: '#dfe1e8',
      base07: '#eff1f5',
      base08: '#bf616a',
      base09: '#d08770',
      base0A: '#ebcb8b',
      base0B: '#a3be8c',
      base0C: '#96b5b4',
      base0D: '#8fa1b3',
      base0E: '#b48ead',
      base0F: '#ab7967',
    },
  },
  {
    id: 'tomorrow-night',
    label: 'Tomorrow Night',
    dark: true,
    colors: {
      base00: '#2d2d2d',
      base01: '#393939',
      base02: '#515151',
      base03: '#999999',
      base04: '#b4b7b4',
      base05: '#cccccc',
      base06: '#e0e0e0',
      base07: '#ffffff',
      base08: '#f2777a',
      base09: '#f99157',
      base0A: '#ffcc66',
      base0B: '#99cc99',
      base0C: '#66cccc',
      base0D: '#6699cc',
      base0E: '#cc99cc',
      base0F: '#a3685a',
    },
  },
];

// Ports of the author's Vim colorschemes. Each value comes from the Vim
// highlight group named in the comment; `tokens` carries what the 16 base16
// slots cannot express (types apart from keywords, bold and italic).
const VIM_PRESETS = [
  {
    // https://github.com/gg582/seoulism.vim (colors/seoulism.vim)
    id: 'seoulism',
    label: 'Seoulism',
    dark: true,
    colors: {
      base00: '#111116', // Normal bg
      base01: '#1a1a22', // bg_alt
      base02: '#2f4fa3', // Visual
      base03: '#7f85ac', // Comment
      base04: '#5f6770', // Delimiter (tag brackets)
      base05: '#b6b5a8', // Normal fg
      base06: '#d7d6d2', // fg_sub
      base07: '#efeeea', // fg_bright
      base08: '#e05a55', // Statement (HTML tag names)
      base09: '#6f8ee6', // Constant
      base0A: '#efeeea', // Type
      base0B: '#e5c15a', // String
      base0C: '#359489', // Special
      base0D: '#6bc0b6', // Function
      base0E: '#e05a55', // Keyword
      base0F: '#6f8ee6', // PreProc
    },
    tokens: {
      comment: { italic: true },
      keyword: { bold: true },
      type: { color: '#efeeea', italic: true }, // Type
      title: { italic: true }, // Structure
      built_in: { color: '#6bc0b6' }, // @function.builtin -> Function
      number: { color: '#e5c15a' }, // Number
      literal: { color: '#f0d487' }, // Boolean
      variable: { color: '#b6b5a8' }, // Identifier
      section: { color: '#e05a55', bold: true }, // Title
      punctuation: { color: '#5f6770' }, // Delimiter
    },
  },
  {
    // https://github.com/gg582/monodraw.vim (colors/monodraw-dark.vim)
    id: 'monodraw-dark',
    label: 'Monodraw Dark',
    dark: true,
    colors: {
      base00: '#0b0d10', // Normal bg
      base01: '#161a20', // CursorLine
      base02: '#2c323c', // Visual
      base03: '#5c6670', // Comment
      base04: '#e5c07b', // htmlTag -> Type
      base05: '#cfd3d6', // Normal fg
      base06: '#abb2bf', // CursorLineNr
      base07: '#ffffff', // PmenuSel fg
      base08: '#e5c07b', // htmlTagName -> Type
      base09: '#84a0c6', // Constant (htmlArg)
      base0A: '#e5c07b', // Type
      base0B: '#8da1b9', // String
      base0C: '#d19a66', // Special
      base0D: '#61afef', // Function
      base0E: '#c678dd', // Keyword
      base0F: '#56b6c2', // PreProc
    },
    tokens: {
      comment: { italic: true },
      keyword: { bold: true },
      type: { color: '#e5c07b' }, // Type
      number: { bold: true },
      literal: { bold: true }, // Boolean
      variable: { color: '#cfd3d6' }, // Identifier
      section: { color: '#61afef', bold: true }, // Title
      operator: { color: '#9fb0c8' }, // Operator
      punctuation: { color: '#7a828a' }, // Delimiter
    },
  },
  {
    // https://github.com/gg582/monodraw.vim (colors/monodraw-light.vim)
    id: 'monodraw-light',
    label: 'Monodraw Light',
    dark: false,
    colors: {
      base00: '#f4f4f4', // Normal bg
      base01: '#e8e8e8', // CursorLine
      base02: '#dde1e6', // Visual
      base03: '#6f6f6f', // Comment
      base04: '#0043ce', // htmlTag -> Type
      base05: '#161616', // Normal fg
      base06: '#393939', // Operator
      base07: '#ffffff', // PmenuSel fg
      base08: '#0043ce', // htmlTagName -> Type
      base09: '#b28600', // Constant (htmlArg)
      base0A: '#0043ce', // Type
      base0B: '#005d5d', // String
      base0C: '#8a3ffc', // Special
      base0D: '#0f62fe', // Function
      base0E: '#d12771', // Keyword
      base0F: '#0072c3', // PreProc
    },
    tokens: {
      comment: { italic: true },
      keyword: { bold: true },
      type: { color: '#0043ce' }, // Type
      number: { bold: true },
      literal: { bold: true }, // Boolean
      variable: { color: '#161616' }, // Identifier
      section: { color: '#0043ce', bold: true }, // Title
      operator: { color: '#393939' }, // Operator
      punctuation: { color: '#525252' }, // Delimiter
    },
  },
];

// The default first, then the Vim ports, then the other base16 schemes.
export const PRESETS = [BASE16_PRESETS[0], ...VIM_PRESETS, ...BASE16_PRESETS.slice(1)];

export const DEFAULT_PRESET_ID = 'github';

export const findPreset = (id) => PRESETS.find((p) => p.id === id);
