package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jinzmedia-atmt/auth"
	"jinzmedia-atmt/models"
	"jinzmedia-atmt/services"
)

type DownloadHandlers struct {
	downloadService *services.DownloadService
}

func NewDownloadHandlers() *DownloadHandlers {
	return &DownloadHandlers{
		downloadService: services.NewDownloadService(),
	}
}

// ListProducts returns available products and user info
func (dh *DownloadHandlers) ListProducts(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("DOWNLOAD ERROR: User not found in context for %s %s", r.Method, r.URL.Path)
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("DOWNLOAD DEBUG: User %s requesting products list", user.Email)

	// Get products and user info
	response, err := dh.downloadService.GetProductsAndUserInfo(user.ID)
	if err != nil {
		log.Printf("DOWNLOAD ERROR: Failed to get products for user %s: %v", user.Email, err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get products: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadProduct serves product files for authenticated users
func (dh *DownloadHandlers) DownloadProduct(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("DOWNLOAD ERROR: User not found in context for %s %s", r.Method, r.URL.Path)
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get URL parameters
	productName := chi.URLParam(r, "product_name")
	platform := chi.URLParam(r, "platform")
	serial := r.URL.Query().Get("serial")

	log.Printf("DOWNLOAD DEBUG: User %s requesting %s/%s with serial %s", user.Email, productName, platform, serial)

	if productName == "" {
		log.Printf("DOWNLOAD ERROR: Missing product name for user %s", user.Email)
		writeErrorResponse(w, http.StatusBadRequest, "Product name is required")
		return
	}

	if platform == "" {
		log.Printf("DOWNLOAD ERROR: Missing platform for user %s", user.Email)
		writeErrorResponse(w, http.StatusBadRequest, "Platform is required")
		return
	}

	if serial == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Serial number is required")
		return
	}

	// Validate product name
	if !models.IsValidProduct(productName) {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid product name")
		return
	}

	// Validate platform
	if platform != "windows" && platform != "macos" {
		log.Printf("DOWNLOAD ERROR: Invalid platform %s for user %s", platform, user.Email)
		writeErrorResponse(w, http.StatusBadRequest, "Invalid platform. Must be 'windows' or 'macos'")
		return
	}

	log.Printf("DOWNLOAD DEBUG: Processing download request for user %s (ID: %s, Owned: %t, Serial: %s)", 
		user.Email, user.ID.Hex(), user.Owned, user.SerialNumber)

	// Process download request
	downloadInfo, err := dh.downloadService.ProcessDownloadRequest(user.ID, productName, platform, serial, r)
	if err != nil {
		log.Printf("DOWNLOAD ERROR: Failed to process download for user %s: %v", user.Email, err)
		// Check specific error types for appropriate HTTP status codes
		switch err.Error() {
		case "user not found":
			writeErrorResponse(w, http.StatusNotFound, "User not found")
		case "user is banned":
			writeErrorResponse(w, http.StatusForbidden, "User account is banned")
		case "you do not own this product":
			writeErrorResponse(w, http.StatusForbidden, "You do not own this product. Please purchase it first.")
		case "serial number does not match":
			writeErrorResponse(w, http.StatusForbidden, "Serial number does not match your account")
		case "file not found":
			writeErrorResponse(w, http.StatusNotFound, "Product file not found")
		default:
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to process download: "+err.Error())
		}
		return
	}

	log.Printf("DOWNLOAD SUCCESS: Serving file %s to user %s", downloadInfo.Filename, user.Email)

	// Serve the file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadInfo.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(downloadInfo.Size, 10))

	http.ServeFile(w, r, downloadInfo.FilePath)
}

// GetDownloadHistory returns user's download history
func (dh *DownloadHandlers) GetDownloadHistory(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("DOWNLOAD ERROR: User not found in context for %s %s", r.Method, r.URL.Path)
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("DOWNLOAD DEBUG: User %s requesting download history", user.Email)

	// Get download history
	downloads, err := dh.downloadService.GetUserDownloadHistory(user.ID)
	if err != nil {
		log.Printf("DOWNLOAD ERROR: Failed to get download history for user %s: %v", user.Email, err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get download history: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"downloads": downloads,
		"total":     len(downloads),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
