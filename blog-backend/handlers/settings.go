package handlers

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/gg582/echo-blog/blog-backend/settings"
)

// GetSettings returns the site settings. Anyone may read them: the frontend
// needs them to style every page.
func (h *Handlers) GetSettings(c *echo.Context) error {
	s, err := h.Settings.Load()
	if err != nil {
		log.Printf("load settings: %v", err)
		return textError(http.StatusInternalServerError, "Error reading settings.")
	}
	return c.JSON(http.StatusOK, s)
}

// SaveSettings validates and stores the site settings.
func (h *Handlers) SaveSettings(c *echo.Context) error {
	var s settings.Settings
	if err := decodeJSON(c, &s); err != nil {
		return textError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if err := s.Validate(); err != nil {
		return textError(http.StatusBadRequest, err.Error())
	}
	if err := h.Settings.Save(s); err != nil {
		log.Printf("save settings: %v", err)
		return textError(http.StatusInternalServerError, "Error saving settings.")
	}
	log.Printf("Settings updated by %s", currentUser(c))
	return c.JSON(http.StatusOK, s)
}
