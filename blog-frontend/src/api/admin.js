// Calls to the admin (dashboard) API. Every call sends the stored bearer
// token; a 401 clears it and sends the user to the login page.
import API_BASE_URL, { authHeaders, clearAuthAndRedirect } from '../config/api';

export class ApiError extends Error {
  constructor(status, message) {
    super(message);
    this.status = status;
  }
}

// request performs an authenticated call and returns the parsed JSON body
// (or null for an empty body). Error responses are plain text on this API,
// so that text becomes the error message.
async function request(path, { method = 'GET', json, body } = {}) {
  const headers = authHeaders(json !== undefined ? { 'Content-Type': 'application/json' } : {});
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    headers,
    body: json !== undefined ? JSON.stringify(json) : body,
  });
  if (response.status === 401) {
    clearAuthAndRedirect();
    throw new ApiError(401, 'Your session has expired. Please log in again.');
  }
  const text = await response.text();
  if (!response.ok) {
    throw new ApiError(response.status, text.trim() || `HTTP error! status: ${response.status}`);
  }
  const data = text ? JSON.parse(text) : null;
  // Uploads answer 202 when only some files were saved.
  return response.status === 202 ? { partial: true, data } : data;
}

export const listFiles = () => request('/api/files');

// Map of file name -> ids of the posts that link to it.
export const fileUsage = () => request('/api/files/usage');

export const deleteFile = (filename) =>
  request('/api/delete-file', { method: 'POST', json: { filename } });

export const renameFile = (from, to, updateReferences) =>
  request('/api/rename-file', { method: 'POST', json: { from, to, updateReferences } });

export const replaceFile = (filename, file) => {
  const form = new FormData();
  form.append('filename', filename);
  form.append('file', file);
  return request('/api/replace-file', { method: 'POST', body: form });
};

// Uploads files under their own names; existing files with the same name are
// overwritten. Returns { partial, data } where data is [{ fileName, url }].
export const uploadFiles = async (files) => {
  const form = new FormData();
  files.forEach((file) => form.append('file', file));
  const result = await request('/api/upload-file', { method: 'POST', body: form });
  return result && result.partial ? result : { partial: false, data: result };
};

export const saveSettings = (settings) =>
  request('/api/settings', { method: 'POST', json: settings });

// Public URL of an uploaded file. version (e.g. the modification time) busts
// the browser cache after a file is replaced.
export const assetUrl = (name, version) => {
  const base = API_BASE_URL || window.location.origin;
  const url = `${base}/assets/${encodeURIComponent(name)}`;
  return version ? `${url}?v=${encodeURIComponent(version)}` : url;
};
