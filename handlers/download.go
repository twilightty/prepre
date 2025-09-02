package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

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
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get products and user info
	response, err := dh.downloadService.GetProductsAndUserInfo(userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get products: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DownloadProduct serves product files for authenticated users
func (dh *DownloadHandlers) DownloadProduct(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get URL parameters
	productName := chi.URLParam(r, "product_name")
	platform := chi.URLParam(r, "platform")
	serial := r.URL.Query().Get("serial")

	if productName == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Product name is required")
		return
	}

	if platform == "" {
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
		writeErrorResponse(w, http.StatusBadRequest, "Invalid platform. Must be 'windows' or 'macos'")
		return
	}

	// Process download request
	downloadInfo, err := dh.downloadService.ProcessDownloadRequest(userID, productName, platform, serial, r)
	if err != nil {
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

	// Serve the file
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadInfo.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(downloadInfo.Size, 10))

	http.ServeFile(w, r, downloadInfo.FilePath)
}

// GetDownloadHistory returns user's download history
func (dh *DownloadHandlers) GetDownloadHistory(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Get download history
	downloads, err := dh.downloadService.GetUserDownloadHistory(userID)
	if err != nil {
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
