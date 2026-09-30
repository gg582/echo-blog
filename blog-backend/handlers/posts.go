package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/gg582/echo-blog/blog-backend/content"
	"github.com/gg582/echo-blog/blog-backend/models"
)

// ListPosts returns every post, rendered.
func (h *Handlers) ListPosts(c *echo.Context) error {
	posts, err := h.Posts.List(c.Request().Context())
	if err != nil {
		return textError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, posts)
}

// GetPost returns one rendered post.
func (h *Handlers) GetPost(c *echo.Context) error {
	post, err := h.Posts.Get(c.Param("id"))
	if errors.Is(err, content.ErrNotFound) {
		return textError(http.StatusNotFound, "Post not found.")
	}
	if err != nil {
		return textError(http.StatusInternalServerError, fmt.Sprintf("Error reading file: %v", err))
	}
	return c.JSON(http.StatusOK, post)
}

// GetRawPost returns the markdown source of one post as plain text.
func (h *Handlers) GetRawPost(c *echo.Context) error {
	raw, err := h.Posts.Raw(c.Param("id"))
	if errors.Is(err, content.ErrNotFound) {
		return textError(http.StatusNotFound, "Post not found.")
	}
	if err != nil {
		return textError(http.StatusInternalServerError, "Error reading file: "+err.Error())
	}
	return c.Blob(http.StatusOK, "text/plain; charset=utf-8", raw)
}

// CreatePost writes a new post from title, author and content. An existing
// post with the same id is overwritten.
func (h *Handlers) CreatePost(c *echo.Context) error {
	slug := c.Param("id")
	if slug == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Post ID (slug) is missing from the URL.",
			"code":    "MISSING_SLUG_IN_URL",
		})
	}

	var req models.NewPostRequest
	if err := decodeJSON(c, &req); err != nil {
		log.Printf("Error decoding new post request body: %v", err)
		return textError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if req.Title == "" || req.Author == "" || req.Content == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Title, author, and content fields in the request body cannot be empty.",
			"code":    "VALIDATION_ERROR",
		})
	}

	if err := h.Posts.EnsureDir(); err != nil {
		log.Printf("Error creating posts directory: %v", err)
		return textError(http.StatusInternalServerError, "Error creating posts directory.")
	}
	if err := h.Posts.Write(slug, content.NewPostMarkdown(req)); err != nil {
		log.Printf("Error saving post file: %v", err)
		return textError(http.StatusInternalServerError, "Error saving post file: "+err.Error())
	}
	log.Printf("New post '%s' (slug: %s) saved", req.Title, slug)

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Post created successfully",
		"id":      slug,
		"url":     "/posts/" + slug,
	})
}

type editPostRequest struct {
	Content string `json:"content"`
}

// EditPost replaces the markdown of a post, creating it if needed.
func (h *Handlers) EditPost(c *echo.Context) error {
	var req editPostRequest
	if err := decodeJSON(c, &req); err != nil {
		return textError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if req.Content == "" {
		return textError(http.StatusBadRequest, "Content cannot be empty.")
	}

	id := c.Param("id")
	if err := h.Posts.Write(id, []byte(req.Content)); err != nil {
		return textError(http.StatusInternalServerError, "Error saving post file: "+err.Error())
	}
	log.Printf("Post %s updated by %s", id, currentUser(c))
	return message(c, http.StatusOK, "Post updated")
}

// DeletePost removes a post.
func (h *Handlers) DeletePost(c *echo.Context) error {
	id := c.Param("id")
	err := h.Posts.Delete(id)
	if errors.Is(err, content.ErrNotFound) {
		return textError(http.StatusNotFound, "Post not found.")
	}
	if err != nil {
		return textError(http.StatusInternalServerError, "Error deleting post file: "+err.Error())
	}
	log.Printf("Post %s deleted by %s", id, currentUser(c))
	return message(c, http.StatusOK, "Post deleted")
}
