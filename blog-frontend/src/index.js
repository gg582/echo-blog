// ~/echo-blog/blog-frontend/src/index.js
import React from 'react';
import ReactDOM from 'react-dom/client';
// Import BrowserRouter for client-side routing
// BrowserRouter requires a 'basename' prop when hosted in a sub-directory on GitHub Pages
import { BrowserRouter } from 'react-router-dom';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    {/* Use BrowserRouter for clean URLs, with basename for GitHub Pages sub-directory */}
    {/* The basename should match your GitHub Pages repository name (e.g., /your-repo-name) */}
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
