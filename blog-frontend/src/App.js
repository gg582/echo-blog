// src/App.js

import React from 'react';
import { Routes, Route } from 'react-router-dom';

// Import all your page components
import HomePage from './pages/HomePage';
import PostDetailPage from './pages/PostDetailPage'; // Assuming you have this
import AboutPage from './pages/AboutPage';
import ContactPage from './pages/ContactPage';
import NewPostPage from './pages/NewPostPage';
import EditPostPage from './pages/EditPostPage';
import DashboardPage from './pages/DashboardPage';
import LoginPage from './pages/LoginPage';

// Import Authentication Context and Protected Route
import { AuthProvider } from './context/AuthContext';
import { SettingsProvider } from './context/SettingsContext';
import ProtectedRoute from './components/ProtectedRoute';

// Import the Header component (now aware of auth state)
import Header from './components/Header'; // Make sure this path is correct

/**
 * Main application content component.
 * This component will be wrapped by AuthProvider and Router to access context and navigation.
 * It contains the global Header and all route definitions.
 */
function AppContent() {
  return (
    <>
      {/* GLOBAL HEADER: This header is fixed and appears on ALL pages */}
      <Header />

      {/* Main content area, pushed down by the fixed header */}
      <div style={{ paddingTop: '60px' }}> {/* Adjust this padding to match header height */}
        {/* ROUTES: The content inside <Routes> changes based on the URL path */}
        <Routes>
          {/* Public routes, accessible to all users */}
          <Route path="/" element={<HomePage />} />
          <Route path="/posts/:id" element={<PostDetailPage />} />
          <Route path="/about" element={<AboutPage />} />
          <Route path="/contact" element={<ContactPage />} />
          <Route path="/login" element={<LoginPage />} />

          {/* Protected route for NewPostPage: only accessible if authenticated */}
          <Route
            path="/new-post"
            element={
              <ProtectedRoute>
                <NewPostPage />
              </ProtectedRoute>
            }
          />

          {/* Protected route for EditPostPage: only accessible if authenticated */}
          <Route
            path="/edit-post/:id"
            element={
              <ProtectedRoute>
                <EditPostPage />
              </ProtectedRoute>
            }
          />

          {/* Protected route for DashboardPage: only accessible if authenticated */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <DashboardPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </div>
    </>
  );
}

/**
 * App Component
 * This is the root component that sets up the Router and AuthProvider for the entire application.
 */
function App() {
  return (
    <AuthProvider>
      <SettingsProvider>
        <AppContent />
      </SettingsProvider>
    </AuthProvider>
  );
}

export default App;
