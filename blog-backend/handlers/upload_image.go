
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gg582/echo-blog/blog-backend/workerpool"
)

// FileJobQueue is the channel used to submit file upload jobs to the worker pool.
var FileJobQueue chan workerpool.UploadJob

// UploadFile handles the upload of one or more files in a single multipart form request.
// It retrieves all uploaded files and submits each one as a job to the worker pool.
func (h *Handlers) UploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form with a 10MB limit.
	r.ParseMultipartForm(10 << 20)

	// Check if any files were uploaded.
	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		http.Error(w, "No files found in the request.", http.StatusBadRequest)
		return
	}

	uploadDirRelative := h.AssetsDir

	// Create the upload directory if it doesn't exist.
	if _, err := os.Stat(uploadDirRelative); os.IsNotExist(err) {
		if err := os.MkdirAll(uploadDirRelative, os.ModePerm); err != nil {
			http.Error(w, fmt.Sprintf("Error creating upload directory '%s': %v", uploadDirRelative, err), http.StatusInternalServerError)
			return
		}
	}

	// Slice to collect results for all successful file uploads.
	results := []map[string]string{}
	allSuccess := true

	// Iterate over all form file fields (key is field name, value is a slice of file headers).
	for fieldName, headers := range r.MultipartForm.File {
		// Iterate over all files uploaded under this form field name.
		for _, handler := range headers {
			file, err := handler.Open()
			if err != nil {
				// Log the error but continue processing other files.
				log.Printf("Error opening file %s from field %s: %v", handler.Filename, fieldName, err)
				allSuccess = false
				continue
			}
			// Close the file after it's been used (i.e., after the worker processes it).
			defer file.Close()

			resultChan := make(chan workerpool.UploadResult)

			job := workerpool.UploadJob{
				File:       file,
				FileHeader: *handler,
				UploadDir:  uploadDirRelative,
				ResultChan: resultChan,
			}

			// Try to submit the job to the worker pool.
			select {
			case FileJobQueue <- job:
				log.Printf("Job for %s submitted to worker pool from field %s.", handler.Filename, fieldName)
			default:
				// If the job queue is full, mark as failure and continue to the next file.
				log.Printf("Worker pool is busy. Failed to submit job for %s.", handler.Filename)
				allSuccess = false
				continue
			}

			// Wait for the result from the worker.
			result := <-resultChan
			close(resultChan)

			if result.Error != nil {
				log.Printf("File upload failed for %s: %v", handler.Filename, result.Error)
				allSuccess = false
				continue
			}

			// Construct the public URL for the saved file based on the current request host/protocol.
			scheme := "http"
			if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
				scheme = "https"
			}
			publicURL := fmt.Sprintf("%s://%s/assets/%s", scheme, r.Host, result.SavedFileName)

			// Append the successful result to the list.
			results = append(results, map[string]string{
				"url":      publicURL,
				"fileName": result.OriginalFileName,
			})
			log.Printf("File %s successfully processed. Public URL: %s", result.OriginalFileName, publicURL)
		}
	}

	// Determine the HTTP status code based on overall success.
	statusCode := http.StatusOK
	if !allSuccess && len(results) == 0 {
		// If all jobs failed.
		http.Error(w, "All file uploads failed. Check server logs for details.", http.StatusInternalServerError)
		return
	} else if !allSuccess {
		// If some files succeeded and some failed, use 202 Accepted to indicate partial success.
		statusCode = http.StatusAccepted
		log.Print("Note: Not all files were processed successfully.")
	}

	// Prepare and send the final JSON response with all successful uploads.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	responseJSON, err := json.Marshal(results)
	if err != nil {
		log.Printf("Error marshalling response JSON: %v", err)
		return
	}
	w.Write(responseJSON)
}
