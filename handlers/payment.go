package handlers

import (
	"encoding/json"
	"jinzmedia-atmt/models"
	"jinzmedia-atmt/services"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

// InitiatePayment creates a new payment session for the authenticated user
func (ph *PaymentHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
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

	// Create payment session
	paymentSession, err := ph.paymentService.InitiatePayment(userID)
	if err != nil {
		if err.Error() == "user not found" {
			writeErrorResponse(w, http.StatusNotFound, "User not found")
			return
		}
		if err.Error() == "user is banned and cannot make payments" {
			writeErrorResponse(w, http.StatusForbidden, "User is banned")
			return
		}
		if err.Error() == "user already owns the product" {
			writeErrorResponse(w, http.StatusConflict, "User already owns the product")
			return
		}
		
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to initiate payment: "+err.Error())
		return
	}

	// Prepare response
	response := models.InitiatePaymentResponse{
		PaymentCode: paymentSession.PaymentCode,
		Amount:      paymentSession.Amount,
		QRImageURL:  paymentSession.QRImageURL,
		ExpiresAt:   paymentSession.ExpiresAt.Format(time.RFC3339),
		Message:     "Please scan the QR code to complete payment. Payment code: " + paymentSession.PaymentCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPaymentStatus retrieves the status of a payment session
func (ph *PaymentHandler) GetPaymentStatus(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")
	sessionID, err := primitive.ObjectIDFromHex(sessionIDStr)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

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

	// Get payment session
	paymentSession, err := ph.paymentService.GetPaymentSession(sessionID)
	if err != nil {
		if err.Error() == "payment session not found" {
			writeErrorResponse(w, http.StatusNotFound, "Payment session not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get payment session: "+err.Error())
		return
	}

	// Check if session belongs to the authenticated user
	if paymentSession.UserID != userID {
		writeErrorResponse(w, http.StatusForbidden, "Unauthorized access to payment session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paymentSession)
}

// GetUserPaymentSessions retrieves all payment sessions for the authenticated user
func (ph *PaymentHandler) GetUserPaymentSessions(w http.ResponseWriter, r *http.Request) {
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

	// Get payment sessions
	sessions, err := ph.paymentService.GetUserPaymentSessions(userID)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get payment sessions: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}
