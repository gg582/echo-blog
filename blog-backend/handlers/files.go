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
