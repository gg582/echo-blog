// ~/echo-blog/blog-frontend/src/pages/PostDetailPage.js

import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import "./PostDetailPage.css";
import API_BASE_URL, { authHeaders, clearAuthAndRedirect } from "../config/api";
import { useAuth } from "../context/AuthContext";
import hljs from "../highlight/hljs";

function PostDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();
  const [post, setPost] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [actionError, setActionError] = useState(null);

  useEffect(() => {
    const fetchPost = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/posts/${id}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({}),
        });
        if (!response.ok) {
          if (response.status === 404) { throw new Error("Post not found."); }
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        setPost(data);
      } catch (e) {
        setError(e);
      } finally {
        setLoading(false);
      }
    };
    fetchPost();
  }, [id]);

  useEffect(() => {
    if (post && post.contentHtml) {
      const postContentElement = document.querySelector('.post-detail-content');
      if (postContentElement) {
        postContentElement.querySelectorAll('pre code').forEach((block) => {
          if (!block.classList.contains('hljs')) {
            hljs.highlightElement(block);
          }
        });
      }
    }
  }, [post]);

  const handleDelete = async () => {
    if (!window.confirm("Delete this post?")) {
      return;
    }
    setActionError(null);
    try {
      const response = await fetch(`${API_BASE_URL}/api/delete-post/${id}`, {
        method: "POST",
        headers: authHeaders(),
      });
      if (response.status === 401) {
        clearAuthAndRedirect();
        return;
      }
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      navigate("/");
    } catch (e) {
      setActionError(e.message);
    }
  };

  if (loading) {
    return (
      <div className="post-detail-page">
        <main className="container">
          <div className="loading-spinner">Loading post...</div>
        </main>
      </div>
    );
  }

  if (error) {
    return (
      <div className="post-detail-page">
        <main className="container">
          <div className="error-box">
            <h2>Error</h2>
            <p>{error.message}</p>
          </div>
        </main>
      </div>
    );
  }

  if (!post) {
    return (
      <div className="post-detail-page">
        <main className="container">
          <div className="error-box">
            <p>Post not found.</p>
          </div>
        </main>
      </div>
    );
  }

  return (
    <div className="post-detail-page">
      <main className="container">
        <h1 className="post-detail-title">{post.title}</h1>
        <p className="post-detail-meta">
          Author: {post.author} | Date:{" "}
          {new Date(post.createdAt).toLocaleDateString()}
        </p>
        <div
          className="post-detail-content"
          dangerouslySetInnerHTML={{ __html: post.contentHtml }}
        ></div>
        {isAuthenticated && (
          <div className="post-actions">
            {actionError && <p className="post-actions-error">{actionError}</p>}
            <button
              type="button"
              className="post-action-btn"
              onClick={() => navigate(`/edit-post/${id}`)}
            >
              Update
            </button>
            <button
              type="button"
              className="post-action-btn post-action-btn-danger"
              onClick={handleDelete}
            >
              Delete
            </button>
          </div>
        )}
      </main>
    </div>
  );
}

export default PostDetailPage;
