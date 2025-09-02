package handlers

import (
	"encoding/json"
	"net/http"

	"jinzmedia-atmt/auth"
	"jinzmedia-atmt/models"
)

type AuthHandlers struct {
	authService *auth.AuthService
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers() *AuthHandlers {
	return &AuthHandlers{
		authService: auth.NewAuthService(),
	}
}

// Register handles user registration
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" || req.FullName == "" || req.SerialNumber == "" {
		writeErrorResponse(w, http.StatusBadRequest, "All fields are required")
		return
	}

	if len(req.Password) < 6 {
		writeErrorResponse(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if err == auth.ErrUserExists {
			writeErrorResponse(w, http.StatusConflict, "User already exists")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	writeSuccessResponse(w, http.StatusCreated, "User created successfully", user)
}

// Login handles user authentication
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	response, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Login failed")
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandlers) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

// GetProfile returns the current user's profile
func (h *AuthHandlers) GetProfile(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	writeJSONResponse(w, http.StatusOK, user)
}

// Logout handles user logout (client-side token removal)
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	writeSuccessResponse(w, http.StatusOK, "Logged out successfully", nil)
}

// writeJSONResponse writes a JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := models.ErrorResponse{
		Error: message,
		Code:  statusCode,
	}
	writeJSONResponse(w, statusCode, response)
}

// writeSuccessResponse writes a success response
func writeSuccessResponse(w http.ResponseWriter, statusCode int, message string, data interface{}) {
	response := models.SuccessResponse{
		Message: message,
		Data:    data,
	}
	writeJSONResponse(w, statusCode, response)
}
