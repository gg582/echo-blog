package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/russross/blackfriday/v2"

	"github.com/gg582/echo-blog/blog-backend/models" // Import models package
	"github.com/gg582/echo-blog/blog-backend/utils"    // Import utils package
)

// GetPostsHandler handles fetching all blog posts.
func (h *Handlers) GetPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := utils.GetPosts(h.PostsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// GetPostByIDHandler handles fetching a single blog post by its ID (slug).
func (h *Handlers) GetPostByIDHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id") // Get the post ID (slug) from the URL

	filePath := filepath.Join(h.PostsDir, postID+".md") // Construct the file path

	content, err := os.ReadFile(filePath) // Use os.ReadFile
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Post not found.", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}

	author, cleanedContent := utils.ParseAuthorAndCleanContent(content) // Use utility function
	htmlContent := blackfriday.Run(cleanedContent)

	title := strings.TrimSuffix(filepath.Base(filePath), ".md") // Default title to filename
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

	post := models.Post{ // Create a Post model instance
		ID:          postID,
		Title:       title,
		ContentHTML: string(htmlContent),
		Author:      author,
		CreatedAt:   createdAt,
		FileName:    postID + ".md",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
