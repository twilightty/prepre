package services

import (
	"jinzmedia-atmt/database"
	"jinzmedia-atmt/models"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type PaymentService struct {
	paymentCollection        *mongo.Collection
	paymentSessionCollection *mongo.Collection
	userCollection          *mongo.Collection
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		paymentCollection:        database.GetCollection("payments"),
		paymentSessionCollection: database.GetCollection("payment_sessions"),
		userCollection:          database.GetCollection("users"),
	}
}

// generatePaymentCode generates a random 8-character alphanumeric code
func (ps *PaymentService) generatePaymentCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 8)
	rand.Read(b)
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}

// InitiatePayment creates a new payment session for the user
func (ps *PaymentService) InitiatePayment(userID primitive.ObjectID) (*models.PaymentSession, error) {
	ctx := context.Background()
	
	// Check if user exists and is not banned
	var user models.User
	err := ps.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.IsBanned {
		return nil, fmt.Errorf("user is banned and cannot make payments")
	}

	if user.Owned {
		return nil, fmt.Errorf("user already owns the product")
	}

	// Generate unique payment code
	var paymentCode string
	for {
		paymentCode = ps.generatePaymentCode()
		// Check if code already exists
		count, err := ps.paymentSessionCollection.CountDocuments(ctx, bson.M{"payment_code": paymentCode})
		if err != nil {
			return nil, fmt.Errorf("failed to check payment code uniqueness: %w", err)
		}
		if count == 0 {
			break
		}
	}

	// Create payment session
	now := time.Now()
	expiresAt := now.Add(15 * time.Minute) // Payment expires in 15 minutes
	
	amount := int64(5000000) // 5,000,000 VND
	
	// Generate QR code URL
	qrImageURL := fmt.Sprintf("https://img.vietqr.io/image/mbbank-28368866886-compact.jpg?amount=%d&addInfo=%s&accountName=%s",
		amount,
		url.QueryEscape("ATMT"+paymentCode),
		url.QueryEscape("NGUYEN HONG QUANG"))

	paymentSession := &models.PaymentSession{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		PaymentCode: paymentCode,
		Amount:      amount,
		Status:      models.PaymentStatusPending,
		QRImageURL:  qrImageURL,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}

	// Save to database
	_, err = ps.paymentSessionCollection.InsertOne(ctx, paymentSession)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment session: %w", err)
	}

	// Update user with payment code (for webhook validation)
	_, err = ps.userCollection.UpdateOne(ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"payment_code": paymentCode, "updated_at": now}})
	if err != nil {
		return nil, fmt.Errorf("failed to update user with payment code: %w", err)
	}

	return paymentSession, nil
}

// ProcessWebhookPayment processes incoming SePay webhook and validates payment
func (ps *PaymentService) ProcessWebhookPayment(webhook *models.SepayWebhookRequest) (*models.Payment, error) {
	ctx := context.Background()
	
	// Convert webhook to payment model
	payment := webhook.ToPayment()
	
	// Check if payment amount is correct (5,000,000 VND)
	if payment.TransferAmount != 5000000 {
		payment.Status = models.PaymentStatusIgnored
		now := time.Now()
		payment.ProcessedAt = &now
		
		// Save ignored payment
		_, err := ps.paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to save ignored payment: %w", err)
		}
		
		log.Printf("Payment ignored - incorrect amount: expected 5000000, got %d", payment.TransferAmount)
		return payment, fmt.Errorf("incorrect payment amount: expected 5000000, got %d", payment.TransferAmount)
	}

	// Extract payment code from content (should contain ATMT<8chars>)
	content := strings.ToUpper(payment.Content)
	atMTIndex := strings.Index(content, "ATMT")
	if atMTIndex == -1 {
		payment.Status = models.PaymentStatusIgnored
		now := time.Now()
		payment.ProcessedAt = &now
		
		// Save ignored payment
		_, err := ps.paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to save ignored payment: %w", err)
		}
		
		log.Printf("Payment ignored - no ATMT code found in content: %s", payment.Content)
		return payment, fmt.Errorf("payment code not found in content: %s", payment.Content)
	}

	// Extract 8-character code after ATMT
	if len(content) < atMTIndex+12 { // ATMT + 8 chars
		payment.Status = models.PaymentStatusIgnored
		now := time.Now()
		payment.ProcessedAt = &now
		
		// Save ignored payment
		_, err := ps.paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to save ignored payment: %w", err)
		}
		
		log.Printf("Payment ignored - invalid code format in content: %s", payment.Content)
		return payment, fmt.Errorf("invalid payment code format in content: %s", payment.Content)
	}

	paymentCode := content[atMTIndex+4 : atMTIndex+12] // Extract 8 chars after ATMT

	// Find user with this payment code
	var user models.User
	err := ps.userCollection.FindOne(ctx, bson.M{"payment_code": paymentCode}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			payment.Status = models.PaymentStatusFailed
			now := time.Now()
			payment.ProcessedAt = &now
			
			// Save failed payment
			_, err := ps.paymentCollection.InsertOne(ctx, payment)
			if err != nil {
				return nil, fmt.Errorf("failed to save failed payment: %w", err)
			}
			
			log.Printf("Payment failed - no user found with payment code: %s", paymentCode)
			return payment, fmt.Errorf("no user found with payment code: %s", paymentCode)
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is banned
	if user.IsBanned {
		payment.Status = models.PaymentStatusFailed
		now := time.Now()
		payment.ProcessedAt = &now
		
		// Save failed payment
		_, err := ps.paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to save failed payment: %w", err)
		}
		
		log.Printf("Payment failed - user is banned: %s", user.Email)
		return payment, fmt.Errorf("user is banned: %s", user.Email)
	}

	// Find and update payment session
	filter := bson.M{
		"payment_code": paymentCode,
		"status":       models.PaymentStatusPending,
	}
	
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":       models.PaymentStatusCompleted,
			"completed_at": now,
		},
	}
	
	result, err := ps.paymentSessionCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment session: %w", err)
	}

	if result.MatchedCount == 0 {
		payment.Status = models.PaymentStatusFailed
		payment.ProcessedAt = &now
		
		// Save failed payment
		_, err := ps.paymentCollection.InsertOne(ctx, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to save failed payment: %w", err)
		}
		
		log.Printf("Payment failed - no pending payment session found for code: %s", paymentCode)
		return payment, fmt.Errorf("no pending payment session found for code: %s", paymentCode)
	}

	// Activate user ownership
	userUpdate := bson.M{
		"$set": bson.M{
			"owned":        true,
			"updated_at":   now,
		},
		"$unset": bson.M{
			"payment_code": "", // Remove payment code after successful payment
		},
	}
	
	_, err = ps.userCollection.UpdateOne(ctx,
		bson.M{"_id": user.ID},
		userUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to activate user ownership: %w", err)
	}

	// Update payment record
	payment.Status = models.PaymentStatusProcessed
	payment.UserID = &user.ID
	payment.ProcessedAt = &now
	
	// Save processed payment
	_, err = ps.paymentCollection.InsertOne(ctx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to save processed payment: %w", err)
	}

	log.Printf("Payment processed successfully for user %s with code %s", user.Email, paymentCode)
	return payment, nil
}

// GetPaymentSession retrieves a payment session by ID
func (ps *PaymentService) GetPaymentSession(sessionID primitive.ObjectID) (*models.PaymentSession, error) {
	ctx := context.Background()
	
	var session models.PaymentSession
	err := ps.paymentSessionCollection.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment session not found")
		}
		return nil, fmt.Errorf("failed to get payment session: %w", err)
	}
	
	return &session, nil
}

// GetUserPaymentSessions retrieves all payment sessions for a user
func (ps *PaymentService) GetUserPaymentSessions(userID primitive.ObjectID) ([]*models.PaymentSession, error) {
	ctx := context.Background()
	
	cursor, err := ps.paymentSessionCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find payment sessions: %w", err)
	}
	defer cursor.Close(ctx)
	
	var sessions []*models.PaymentSession
	for cursor.Next(ctx) {
		var session models.PaymentSession
		if err := cursor.Decode(&session); err != nil {
			return nil, fmt.Errorf("failed to decode payment session: %w", err)
		}
		sessions = append(sessions, &session)
	}
	
	return sessions, nil
}

// GetUserById retrieves a user by their ID (for checking updated ownership status)
func (ps *PaymentService) GetUserById(userID primitive.ObjectID) (*models.User, error) {
	ctx := context.Background()
	
	var user models.User
	err := ps.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}
