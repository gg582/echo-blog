// Package handlers implements the blog's HTTP API on top of Echo.
//
// Error responses keep the plain-text format of net/http's http.Error: a
// handler returns textError(code, msg) and the server's error handler writes
// msg as text/plain.
package handlers

import (
	"encoding/json"

	"github.com/labstack/echo/v5"

	"github.com/gg582/echo-blog/blog-backend/assets"
	"github.com/gg582/echo-blog/blog-backend/auth"
	"github.com/gg582/echo-blog/blog-backend/content"
)

// Handlers holds the dependencies of the HTTP handlers.
type Handlers struct {
	Posts   *content.Posts
	About   content.Page
	Contact content.Page
	Assets  *assets.Store
	Users   *auth.Users
	Tokens  *auth.Signer
}

// Register mounts the API routes on e.
func (h *Handlers) Register(e *echo.Echo) {
	requireAuth := h.RequireAuth

	e.POST("/api/login", h.Login)

	e.POST("/api/posts", h.ListPosts)
	e.POST("/api/posts/:id", h.GetPost)
	e.GET("/api/posts/:id/raw", h.GetRawPost)
	e.POST("/api/new-post/:id", h.CreatePost)
	e.POST("/api/edit-post/:id", h.EditPost, requireAuth)
	e.POST("/api/delete-post/:id", h.DeletePost, requireAuth)

	e.GET("/api/about", h.GetAbout)
	e.GET("/api/contact", h.GetContact)

	e.GET("/api/files", h.ListFiles, requireAuth)
	e.POST("/api/delete-file", h.DeleteFile, requireAuth)
	e.POST("/api/upload-file", h.UploadFiles)
}

// textError returns an error that the server renders as a plain-text body.
func textError(code int, msg string) error {
	return echo.NewHTTPError(code, msg)
}

// decodeJSON decodes the request body into v with encoding/json, so decode
// errors read the same as before (e.g. "unexpected EOF").
func decodeJSON(c *echo.Context, v any) error {
	return json.NewDecoder(c.Request().Body).Decode(v)
}

// message responds with {"message": msg}.
func message(c *echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"message": msg})
}
