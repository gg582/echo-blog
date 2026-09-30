package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gg582/echo-blog/blog-backend/models"
)

// GetRawPostHandler returns the raw markdown of a single post as plain text.
func (h *Handlers) GetRawPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")

	content, err := os.ReadFile(filepath.Join(h.PostsDir, postID+".md"))
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Post not found.", http.StatusNotFound)
			return
		}
		http.Error(w, "Error reading file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// EditPostRequest is the JSON body expected by EditPostHandler.
type EditPostRequest struct {
	Content string `json:"content"`
}

// EditPostHandler creates or overwrites posts/{id}.md with the given markdown.
func (h *Handlers) EditPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")

	var req EditPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Content == "" {
		http.Error(w, "Content cannot be empty.", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.PostsDir, postID+".md")
	if err := os.WriteFile(filePath, []byte(req.Content), 0644); err != nil {
		http.Error(w, "Error saving post file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post updated"})
}

// DeletePostHandler deletes posts/{id}.md.
func (h *Handlers) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "id")

	filePath := filepath.Join(h.PostsDir, postID+".md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Post not found.", http.StatusNotFound)
		return
	}
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Error deleting post file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted"})
}

// GetFilesHandler lists the files in the assets directory, newest first.
func (h *Handlers) GetFilesHandler(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(h.AssetsDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.FileInfo{})
			return
		}
		http.Error(w, "Error reading assets directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	files := []models.FileInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, models.FileInfo{
			Name:      info.Name(),
			Size:      info.Size(),
			ModifiedAt: info.ModTime().UTC().Format(time.RFC3339),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModifiedAt > files[j].ModifiedAt
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// DeleteFileRequest is the JSON body expected by DeleteFileHandler.
type DeleteFileRequest struct {
	Filename string `json:"filename"`
}

// DeleteFileHandler deletes a file from the assets directory.
func (h *Handlers) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	filename := req.Filename
	if filename == "" ||
		strings.ContainsAny(filename, `/\`) ||
		strings.Contains(filename, "..") {
		http.Error(w, "Invalid filename.", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(h.AssetsDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Error deleting file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "File deleted"})
}
