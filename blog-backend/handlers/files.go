package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"

	"github.com/gg582/echo-blog/blog-backend/assets"
)

// maxUploadMemory is how much of a multipart upload is kept in memory; the
// rest is spilled to temporary files.
const maxUploadMemory = 10 << 20

// ListFiles returns the uploaded files, newest first.
func (h *Handlers) ListFiles(c *echo.Context) error {
	files, err := h.Assets.List()
	if err != nil {
		return textError(http.StatusInternalServerError, "Error reading assets directory: "+err.Error())
	}
	return c.JSON(http.StatusOK, files)
}

type deleteFileRequest struct {
	Filename string `json:"filename"`
}

// DeleteFile removes one uploaded file.
func (h *Handlers) DeleteFile(c *echo.Context) error {
	var req deleteFileRequest
	if err := decodeJSON(c, &req); err != nil {
		return textError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	err := h.Assets.Delete(req.Filename)
	switch {
	case errors.Is(err, assets.ErrInvalidName):
		return textError(http.StatusBadRequest, "Invalid filename.")
	case errors.Is(err, assets.ErrNotFound):
		return textError(http.StatusNotFound, "File not found.")
	case err != nil:
		return textError(http.StatusInternalServerError, "Error deleting file: "+err.Error())
	}
	log.Printf("File %s deleted by %s", req.Filename, currentUser(c))
	return message(c, http.StatusOK, "File deleted")
}

// uploadResult fields are ordered to match the previous map-based output.
type uploadResult struct {
	FileName string `json:"fileName"`
	URL      string `json:"url"`
}

// UploadFiles saves every file of a multipart form and returns their public
// URLs. It responds 200 if all succeeded, 202 if only some did, and 500 if
// none did.
func (h *Handlers) UploadFiles(c *echo.Context) error {
	req := c.Request()
	req.ParseMultipartForm(maxUploadMemory)
	if req.MultipartForm == nil || len(req.MultipartForm.File) == 0 {
		return textError(http.StatusBadRequest, "No files found in the request.")
	}

	if err := h.Assets.EnsureDir(); err != nil {
		return textError(http.StatusInternalServerError,
			fmt.Sprintf("Error creating upload directory '%s': %v", h.Assets.Dir(), err))
	}

	baseURL := publicBaseURL(req)
	results := []uploadResult{}
	allSuccess := true
	for field, headers := range req.MultipartForm.File {
		for _, fh := range headers {
			saved, err := h.Assets.Save(req.Context(), fh)
			if err != nil {
				log.Printf("File upload failed for %s (field %s): %v", fh.Filename, field, err)
				allSuccess = false
				continue
			}
			publicURL := baseURL + "/assets/" + saved
			results = append(results, uploadResult{URL: publicURL, FileName: fh.Filename})
			log.Printf("File %s successfully processed. Public URL: %s", fh.Filename, publicURL)
		}
	}

	status := http.StatusOK
	if !allSuccess {
		if len(results) == 0 {
			return textError(http.StatusInternalServerError, "All file uploads failed. Check server logs for details.")
		}
		status = http.StatusAccepted
		log.Print("Note: Not all files were processed successfully.")
	}

	// json.Marshal (no trailing newline) keeps the body byte-identical to before.
	body, err := json.Marshal(results)
	if err != nil {
		return err
	}
	return c.JSONBlob(status, body)
}

// publicBaseURL returns scheme://host of the request as the client sees it,
// honouring X-Forwarded-Proto from a TLS-terminating proxy.
func publicBaseURL(req *http.Request) string {
	u := url.URL{Scheme: "http", Host: req.Host}
	if req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") {
		u.Scheme = "https"
	}
	return u.String()
}

// FileUsage returns, for every uploaded file that posts link to, the ids of
// those posts: {"name.png": ["post-id", ...]}.
func (h *Handlers) FileUsage(c *echo.Context) error {
	files, err := h.Assets.List()
	if err != nil {
		return textError(http.StatusInternalServerError, "Error reading assets directory: "+err.Error())
	}
	sources, err := h.Posts.Sources(c.Request().Context())
	if err != nil {
		return textError(http.StatusInternalServerError, err.Error())
	}

	usage := map[string][]string{}
	for _, f := range files {
		for _, src := range sources {
			if assets.Mentions(src.Markdown, f.Name) {
				usage[f.Name] = append(usage[f.Name], src.ID)
			}
		}
	}
	return c.JSON(http.StatusOK, usage)
}

// ReplaceFile overwrites an existing file with the uploaded "file" field,
// keeping the name given in the "filename" field so links stay valid.
func (h *Handlers) ReplaceFile(c *echo.Context) error {
	name := c.FormValue("filename")
	fh, err := c.FormFile("file")
	if err != nil {
		return textError(http.StatusBadRequest, "No file found in the request.")
	}
	file, err := fh.Open()
	if err != nil {
		return textError(http.StatusBadRequest, "Error reading uploaded file: "+err.Error())
	}
	defer file.Close()

	info, err := h.Assets.Replace(name, file)
	if err := assetError(err); err != nil {
		return err
	}
	log.Printf("File %s replaced by %s (%d bytes)", name, currentUser(c), info.Size)
	return c.JSON(http.StatusOK, map[string]any{"message": "File replaced", "file": info})
}

type renameFileRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	// UpdateReferences also rewrites links to the file in posts.
	UpdateReferences bool `json:"updateReferences"`
}

// RenameFile renames an uploaded file and optionally rewrites the links to it
// in every post.
func (h *Handlers) RenameFile(c *echo.Context) error {
	var req renameFileRequest
	if err := decodeJSON(c, &req); err != nil {
		return textError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	info, err := h.Assets.Rename(req.From, req.To)
	if err := assetError(err); err != nil {
		return err
	}
	log.Printf("File %s renamed to %s by %s", req.From, req.To, currentUser(c))

	updated := []string{}
	if req.UpdateReferences && req.From != req.To {
		updated, err = h.rewriteLinks(c, req.From, req.To)
		if err != nil {
			return textError(http.StatusInternalServerError,
				fmt.Sprintf("File renamed, but updating posts failed after %v: %v", updated, err))
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"message":      "File renamed",
		"file":         info,
		"updatedPosts": updated,
	})
}

// rewriteLinks points every post's links to from at to, and returns the ids
// of the posts it changed.
func (h *Handlers) rewriteLinks(c *echo.Context, from, to string) ([]string, error) {
	sources, err := h.Posts.Sources(c.Request().Context())
	if err != nil {
		return nil, err
	}
	updated := []string{}
	for _, src := range sources {
		markdown, changed := assets.RewriteLinks(src.Markdown, from, to)
		if !changed {
			continue
		}
		if err := h.Posts.Write(src.ID, markdown); err != nil {
			return updated, fmt.Errorf("post %s: %w", src.ID, err)
		}
		updated = append(updated, src.ID)
	}
	log.Printf("Links to %s rewritten to %s in posts %v", from, to, updated)
	return updated, nil
}

// assetError maps assets store errors to HTTP errors; nil stays nil.
func assetError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, assets.ErrInvalidName):
		return textError(http.StatusBadRequest, "Invalid filename.")
	case errors.Is(err, assets.ErrNotFound):
		return textError(http.StatusNotFound, "File not found.")
	case errors.Is(err, assets.ErrExists):
		return textError(http.StatusConflict, "A file with that name already exists.")
	default:
		return textError(http.StatusInternalServerError, "Error updating file: "+err.Error())
	}
}
