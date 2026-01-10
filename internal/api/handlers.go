package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/secure-storage/internal/storage"
)

// Handler contains HTTP handlers for the API
type Handler struct {
	storage *storage.Service
}

// NewHandler creates a new handler with the given storage service
func NewHandler(storage *storage.Service) *Handler {
	return &Handler{
		storage: storage,
	}
}

// response is a generic JSON response structure
type response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// sendJSON helper to write JSON responses
func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// UploadHandler handles POST /upload requests
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, response{
			Success: false,
			Error:   "method not allowed, use POST",
		})
		return
	}

	// Parse the multipart form with a 10MB limit
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		sendJSON(w, http.StatusBadRequest, response{
			Success: false,
			Error:   "failed to parse form data",
		})
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting file from form: %v", err)
		sendJSON(w, http.StatusBadRequest, response{
			Success: false,
			Error:   "no file provided in 'file' field",
		})
		return
	}
	defer file.Close()

	// Save and encrypt the file
	filename := header.Filename
	if err := h.storage.SaveFile(filename, file); err != nil {
		log.Printf("Error saving file %s: %v", filename, err)
		sendJSON(w, http.StatusBadRequest, response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	log.Printf("Successfully uploaded and encrypted: %s", filename)
	sendJSON(w, http.StatusOK, response{
		Success: true,
		Message: "file uploaded and encrypted successfully",
	})
}

// DownloadHandler handles GET /download/{filename} requests
func (h *Handler) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, response{
			Success: false,
			Error:   "method not allowed, use GET",
		})
		return
	}

	// Extract filename from URL path
	// Path is like /download/myfile.txt
	path := r.URL.Path
	prefix := "/download/"
	if !strings.HasPrefix(path, prefix) {
		sendJSON(w, http.StatusBadRequest, response{
			Success: false,
			Error:   "invalid path",
		})
		return
	}

	filename := strings.TrimPrefix(path, prefix)
	if filename == "" {
		sendJSON(w, http.StatusBadRequest, response{
			Success: false,
			Error:   "filename is required",
		})
		return
	}

	// Load and decrypt the file
	data, err := h.storage.LoadFile(filename)
	if err != nil {
		log.Printf("Error loading file %s: %v", filename, err)

		// Return 404 for not found, 400 for other errors
		if strings.Contains(err.Error(), "not found") {
			sendJSON(w, http.StatusNotFound, response{
				Success: false,
				Error:   "file not found",
			})
		} else {
			sendJSON(w, http.StatusBadRequest, response{
				Success: false,
				Error:   err.Error(),
			})
		}
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")

	// Stream the decrypted file back
	w.Write(data)
	log.Printf("Successfully downloaded: %s", filename)
}

// HealthHandler handles GET /health requests for container health checks
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, response{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	sendJSON(w, http.StatusOK, response{
		Success: true,
		Message: "healthy",
	})
}
