package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday/v2"

	"github.com/gg582/echo-blog/blog-backend/models"
	"github.com/gg582/echo-blog/blog-backend/utils"
)

// Handlers groups the HTTP handlers that depend on filesystem paths.
type Handlers struct {
	PostsDir  string
	AssetsDir string
	AboutMD   string
	ContactMD string
}

// NewHandlers creates a Handlers with the given content paths.
func NewHandlers(postsDir, assetsDir, aboutMD, contactMD string) *Handlers {
	return &Handlers{
		PostsDir:  postsDir,
		AssetsDir: assetsDir,
		AboutMD:   aboutMD,
		ContactMD: contactMD,
	}
}

// GetAboutPageHandler handles fetching the content for the about page.
func (h *Handlers) GetAboutPageHandler(w http.ResponseWriter, r *http.Request) {
	filePath := h.AboutMD

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "About page content not found.", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error reading about file: %v", err), http.StatusInternalServerError)
		return
	}

	author, cleanedContent := utils.ParseAuthorAndCleanContent(content)
	htmlContent := blackfriday.Run(cleanedContent)

	// Determine title from cleaned content, or use a default
	title := "About Us"
	lines := strings.Split(string(cleanedContent), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			if strings.HasPrefix(trimmedLine, "#") {
				title = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "#"))
			} else {
				title = trimmedLine
			}
			break
		}
	}

	fileInfo, _ := os.Stat(filePath)
	createdAt := fileInfo.ModTime()

	post := models.Post{
		ID:          "about",
		Title:       title,
		ContentHTML: string(htmlContent),
		Author:      author,
		CreatedAt:   createdAt,
		FileName:    filepath.Base(filePath),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

// GetContactPageHandler handles fetching the content for the contact page.
func (h *Handlers) GetContactPageHandler(w http.ResponseWriter, r *http.Request) {
	filePath := h.ContactMD

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Contact page content not found.", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error reading contact file: %v", err), http.StatusInternalServerError)
		return
	}

	author, cleanedContent := utils.ParseAuthorAndCleanContent(content)
	htmlContent := blackfriday.Run(cleanedContent)

	// Determine title from cleaned content, or use a default
	title := "Contact Us"
	lines := strings.Split(string(cleanedContent), "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			if strings.HasPrefix(trimmedLine, "#") {
				title = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "#"))
			} else {
				title = trimmedLine
			}
			break
		}
	}

	fileInfo, _ := os.Stat(filePath)
	createdAt := fileInfo.ModTime()

	post := models.Post{
		ID:          "contact",
		Title:       title,
		ContentHTML: string(htmlContent),
		Author:      author,
		CreatedAt:   createdAt,
		FileName:    filepath.Base(filePath),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
