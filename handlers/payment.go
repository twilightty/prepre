package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"jinzmedia-atmt/auth"
	"jinzmedia-atmt/models"
	"jinzmedia-atmt/services"
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
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("PAYMENT ERROR: User not found in context for payment initiation")
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("PAYMENT DEBUG: User %s (ID: %s) initiating payment", user.Email, user.ID.Hex())

	// Create payment session
	paymentSession, err := ph.paymentService.InitiatePayment(user.ID)
	if err != nil {
		log.Printf("PAYMENT ERROR: Failed to initiate payment for user %s: %v", user.Email, err)
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

	log.Printf("PAYMENT SUCCESS: Payment session created for user %s with code %s", user.Email, paymentSession.PaymentCode)

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
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("PAYMENT ERROR: User not found in context for payment status check")
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("PAYMENT DEBUG: User %s checking payment status for session %s", user.Email, sessionIDStr)

	// Get payment session
	paymentSession, err := ph.paymentService.GetPaymentSession(sessionID)
	if err != nil {
		log.Printf("PAYMENT ERROR: Failed to get payment session %s: %v", sessionIDStr, err)
		if err.Error() == "payment session not found" {
			writeErrorResponse(w, http.StatusNotFound, "Payment session not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get payment session: "+err.Error())
		return
	}

	// Check if session belongs to the authenticated user
	if paymentSession.UserID != user.ID {
		log.Printf("PAYMENT ERROR: User %s tried to access payment session %s belonging to another user", user.Email, sessionIDStr)
		writeErrorResponse(w, http.StatusForbidden, "Unauthorized access to payment session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paymentSession)
}

// GetUserPaymentSessions retrieves all payment sessions for the authenticated user
func (ph *PaymentHandler) GetUserPaymentSessions(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("PAYMENT ERROR: User not found in context for payment sessions request")
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("PAYMENT DEBUG: User %s requesting payment sessions", user.Email)

	// Get payment sessions
	sessions, err := ph.paymentService.GetUserPaymentSessions(user.ID)
	if err != nil {
		log.Printf("PAYMENT ERROR: Failed to get payment sessions for user %s: %v", user.Email, err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get payment sessions: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// RefreshPayment handles "I have paid" button - checks if user now has access
func (ph *PaymentHandler) RefreshPayment(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		log.Printf("PAYMENT ERROR: User not found in context for payment refresh")
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("PAYMENT DEBUG: User %s clicked 'I have paid' - checking ownership status", user.Email)

	// Check user's current ownership status by refreshing from database
	updatedUser, err := ph.paymentService.GetUserById(user.ID)
	if err != nil {
		log.Printf("PAYMENT ERROR: Failed to get updated user data for %s: %v", user.Email, err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to check account status")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"owned":   updatedUser.Owned,
		"reload":  true,
	}

	if updatedUser.Owned {
		response["message"] = "Payment confirmed! You now have access to all products."
		response["redirect"] = "/dashboard"
		log.Printf("PAYMENT SUCCESS: User %s payment confirmed - granting access", user.Email)
	} else {
		response["message"] = "Payment not yet confirmed. Please wait a moment and try again."
		response["redirect"] = "/payment"
		log.Printf("PAYMENT PENDING: User %s payment still pending", user.Email)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
