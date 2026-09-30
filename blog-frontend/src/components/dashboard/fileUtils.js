// Helpers shared by the dashboard file manager.

const KINDS = {
  image: ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp', 'ico', 'avif'],
  video: ['mp4', 'webm', 'ogg', 'mov', 'mkv'],
  audio: ['mp3', 'wav', 'flac', 'm4a', 'aac', 'oga'],
};

export const extension = (name) => {
  const dot = name.lastIndexOf('.');
  return dot > 0 ? name.slice(dot + 1).toLowerCase() : '';
};

// 'image', 'video', 'audio' or 'file'.
export const fileKind = (name) => {
  const ext = extension(name);
  return Object.keys(KINDS).find((kind) => KINDS[kind].includes(ext)) || 'file';
};

// Format a byte count as B, KB or MB for display.
export const formatSize = (bytes) => {
  if (typeof bytes !== 'number' || isNaN(bytes)) { return '-'; }
  if (bytes >= 1024 * 1024) { return `${(bytes / (1024 * 1024)).toFixed(1)} MB`; }
  if (bytes >= 1024) { return `${(bytes / 1024).toFixed(1)} KB`; }
  return `${bytes} B`;
};

export const formatDate = (iso) => (iso ? new Date(iso).toLocaleString() : '-');

// The markdown snippet the post editor inserts for an uploaded file.
export const markdownFor = (name, url) => {
  switch (fileKind(name)) {
    case 'image':
      return `![${name}](${url})`;
    case 'video':
      return `<video controls width="600">\n  <source src="${url}">\n</video>`;
    case 'audio':
      return `<audio controls>\n  <source src="${url}">\n</audio>`;
    default:
      return `[Download File: ${name}](${url})`;
  }
};

// Same rules as the backend: a plain name that cannot leave the directory.
export const validFileName = (name) =>
  name.trim() !== '' && !/[/\\]/.test(name) && !name.includes('..');

export const copyText = async (text) => {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text);
    return;
  }
  // Fallback for plain-HTTP origins where the Clipboard API is unavailable.
  const area = document.createElement('textarea');
  area.value = text;
  area.style.position = 'fixed';
  area.style.opacity = '0';
  document.body.appendChild(area);
  area.select();
  document.execCommand('copy');
  document.body.removeChild(area);
};

export const SORTS = {
  newest: { label: 'Newest first', compare: (a, b) => b.modifiedAt.localeCompare(a.modifiedAt) },
  oldest: { label: 'Oldest first', compare: (a, b) => a.modifiedAt.localeCompare(b.modifiedAt) },
  name: { label: 'Name', compare: (a, b) => a.name.localeCompare(b.name) },
  largest: { label: 'Largest first', compare: (a, b) => b.size - a.size },
};

export const FILTERS = {
  all: { label: 'All files', match: () => true },
  image: { label: 'Images', match: (f) => fileKind(f.name) === 'image' },
  media: { label: 'Video & audio', match: (f) => ['video', 'audio'].includes(fileKind(f.name)) },
  other: { label: 'Other files', match: (f) => fileKind(f.name) === 'file' },
  unused: { label: 'Not used in any post', match: (f, usage) => !(usage[f.name] || []).length },
};
