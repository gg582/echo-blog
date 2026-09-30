package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/gg582/echo-blog/blog-backend/content"
)

// GetAbout returns the rendered about page.
func (h *Handlers) GetAbout(c *echo.Context) error {
	return renderPage(c, h.About, "About", "about")
}

// GetContact returns the rendered contact page.
func (h *Handlers) GetContact(c *echo.Context) error {
	return renderPage(c, h.Contact, "Contact", "contact")
}

// renderPage responds with page rendered as a Post. label and file name the
// page in error messages ("About page content not found.").
func renderPage(c *echo.Context, page content.Page, label, file string) error {
	post, err := page.Load()
	if errors.Is(err, content.ErrNotFound) {
		return textError(http.StatusNotFound, label+" page content not found.")
	}
	if err != nil {
		return textError(http.StatusInternalServerError, fmt.Sprintf("Error reading %s file: %v", file, err))
	}
	return c.JSON(http.StatusOK, post)
}
