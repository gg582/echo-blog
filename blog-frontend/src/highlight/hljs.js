import hljs from 'highlight.js';
import freedesktop from './languages/freedesktop';
// Colors come from CSS variables so the dashboard can change them at runtime.
import './theme.css';

// Custom language from the gg582/highlight.js fork. Blog posts also use the
// "desktop" fence name for XDG shortcut files.
hljs.registerLanguage('freedesktop', freedesktop);
hljs.registerAliases('desktop', { languageName: 'freedesktop' });

export default hljs;
