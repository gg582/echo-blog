package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5" // Import chi for URLParam

	"github.com/gg582/echo-blog/blog-backend/models"
)

// CreateNewPostHandler handles the submission of a new blog post.
// It expects a JSON payload with title, author, and markdown content.
// The post ID (slug) is now provided in the URL path by the frontend.
func (h *Handlers) CreateNewPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get the postSlug directly from the URL path.
	postSlug := chi.URLParam(r, "id") // Assuming your route is /api/new-post/{id}
	if postSlug == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Post ID (slug) is missing from the URL.",
			"code":    "MISSING_SLUG_IN_URL",
		})
		return
	}

	var newPost models.NewPostRequest // Using the defined struct for request body
	
	// Decode the JSON request body
	err := json.NewDecoder(r.Body).Decode(&newPost)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		log.Printf("Error decoding new post request body: %v", err)
		return
	}

	// Basic validation for fields from the request body
	if newPost.Title == "" || newPost.Author == "" || newPost.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Title, author, and content fields in the request body cannot be empty.",
			"code":    "VALIDATION_ERROR",
		})
		return
	}

	// Define the directory for posts
	postsDir := h.PostsDir
	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		err = os.MkdirAll(postsDir, os.ModePerm)
		if err != nil {
			http.Error(w, "Error creating posts directory.", http.StatusInternalServerError)
			log.Printf("Error creating posts directory: %v", err)
			return
		}
	}

	// Construct the filename using the slug from the URL
	filename := postSlug + ".md"
	filePath := filepath.Join(postsDir, filename)

    // JUST OVERWRITE IT
	// // Check if a file with this slug already exists to prevent overwriting
	// if _, err := os.Stat(filePath); err == nil {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusConflict) // 409 Conflict if resource already exists
	// 	json.NewEncoder(w).Encode(map[string]string{
	// 		"message": fmt.Sprintf("A post with the ID '%s' (from title: '%s') already exists.", postSlug, newPost.Title),
	// 		"code":    "DUPLICATE_SLUG",
	// 		"slug":    postSlug,
	// 	})
	// 	return
	// }

	// Prepare the content to be written to the markdown file
	// Include the author as front matter. You might also want to include the original title here.
	markdownContent := fmt.Sprintf("---\nauthor: %s\n---\n\n# %s\n\n%s", newPost.Author, newPost.Title, newPost.Content)

	// Write the markdown content to the file
	err = os.WriteFile(filePath, []byte(markdownContent), 0644)
	if err != nil {
		http.Error(w, "Error saving post file: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error saving post file: %v", err)
		return
	}

	log.Printf("New post '%s' (slug: %s) saved to %s", newPost.Title, postSlug, filePath)

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created for successful resource creation
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post created successfully",
		"id":      postSlug,
		"url":     "/posts/" + postSlug,
	})
}
