package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"jinzmedia-atmt/models"
	"jinzmedia-atmt/services"
)

type WebhookHandler struct {
	paymentService *services.PaymentService
}

func NewWebhookHandler(paymentService *services.PaymentService) *WebhookHandler {
	return &WebhookHandler{
		paymentService: paymentService,
	}
}

// HandleSepayWebhook handles webhook calls from SePay
func (h *WebhookHandler) HandleSepayWebhook(w http.ResponseWriter, r *http.Request) {
	// Validate authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "ApiKey xoxoxoxoxoxo" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse webhook payload
	var webhookReq models.SepayWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&webhookReq); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Process the payment
	_, err := h.paymentService.ProcessWebhookPayment(&webhookReq)
	if err != nil {
		// Log error but return success to prevent webhook retries
		// In production, you might want to implement proper error handling
		// and return appropriate status codes based on error type
		if strings.Contains(err.Error(), "payment already processed") {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "already_processed"})
			return
		}
		
		// For other errors, still return success to prevent webhook spam
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
