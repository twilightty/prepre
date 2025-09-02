package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Payment represents a payment transaction from SePay
type Payment struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	SepayID           int64              `bson:"sepay_id" json:"id"`                          // ID giao dịch trên SePay
	Gateway           string             `bson:"gateway" json:"gateway"`                      // Brand name của ngân hàng
	TransactionDate   string             `bson:"transaction_date" json:"transactionDate"`    // Thời gian xảy ra giao dịch phía ngân hàng
	AccountNumber     string             `bson:"account_number" json:"accountNumber"`        // Số tài khoản ngân hàng
	Code              *string            `bson:"code" json:"code"`                           // Mã code thanh toán (có thể null)
	Content           string             `bson:"content" json:"content"`                     // Nội dung chuyển khoản
	TransferType      string             `bson:"transfer_type" json:"transferType"`          // Loại giao dịch. in là tiền vào, out là tiền ra
	TransferAmount    int64              `bson:"transfer_amount" json:"transferAmount"`      // Số tiền giao dịch
	Accumulated       int64              `bson:"accumulated" json:"accumulated"`             // Số dư tài khoản (lũy kế)
	SubAccount        *string            `bson:"sub_account" json:"subAccount"`              // Tài khoản ngân hàng phụ (có thể null)
	ReferenceCode     string             `bson:"reference_code" json:"referenceCode"`       // Mã tham chiếu của tin nhắn sms
	Description       string             `bson:"description" json:"description"`            // Toàn bộ nội dung tin nhắn sms
	ProcessedAt       *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"` // Thời gian xử lý webhook
	Status            PaymentStatus      `bson:"status" json:"status"`                       // Trạng thái xử lý
	UserID            *primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"` // ID người dùng (nếu xác định được)
	ProductID         *primitive.ObjectID `bson:"product_id,omitempty" json:"product_id,omitempty"` // ID sản phẩm (nếu xác định được)
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

// PaymentStatus represents the status of payment processing
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"   // Chờ xử lý
	PaymentStatusProcessed PaymentStatus = "processed" // Đã xử lý thành công
	PaymentStatusCompleted PaymentStatus = "completed" // Hoàn thành (cho payment session)
	PaymentStatusExpired   PaymentStatus = "expired"   // Hết hạn (cho payment session)
	PaymentStatusFailed    PaymentStatus = "failed"    // Xử lý thất bại
	PaymentStatusIgnored   PaymentStatus = "ignored"   // Bỏ qua (không liên quan đến hệ thống)
)

// SepayWebhookRequest represents the incoming webhook request from SePay
type SepayWebhookRequest struct {
	ID              int64   `json:"id"`
	Gateway         string  `json:"gateway"`
	TransactionDate string  `json:"transactionDate"`
	AccountNumber   string  `json:"accountNumber"`
	Code            *string `json:"code"`
	Content         string  `json:"content"`
	TransferType    string  `json:"transferType"`
	TransferAmount  int64   `json:"transferAmount"`
	Accumulated     int64   `json:"accumulated"`
	SubAccount      *string `json:"subAccount"`
	ReferenceCode   string  `json:"referenceCode"`
	Description     string  `json:"description"`
}

// ToPayment converts SepayWebhookRequest to Payment model
func (s *SepayWebhookRequest) ToPayment() *Payment {
	now := time.Now()
	return &Payment{
		SepayID:         s.ID,
		Gateway:         s.Gateway,
		TransactionDate: s.TransactionDate,
		AccountNumber:   s.AccountNumber,
		Code:            s.Code,
		Content:         s.Content,
		TransferType:    s.TransferType,
		TransferAmount:  s.TransferAmount,
		Accumulated:     s.Accumulated,
		SubAccount:      s.SubAccount,
		ReferenceCode:   s.ReferenceCode,
		Description:     s.Description,
		ProcessedAt:     &now,
		Status:          PaymentStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// PaymentSession represents a payment session for QR code generation
type PaymentSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	PaymentCode  string             `bson:"payment_code" json:"payment_code"`
	Amount       int64              `bson:"amount" json:"amount"`
	Status       PaymentStatus      `bson:"status" json:"status"`
	QRImageURL   string             `bson:"qr_image_url" json:"qr_image_url"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
	CompletedAt  *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// InitiatePaymentRequest represents the request to initiate a payment
type InitiatePaymentRequest struct {
	UserID primitive.ObjectID `json:"user_id"`
}

// InitiatePaymentResponse represents the response of payment initiation
type InitiatePaymentResponse struct {
	PaymentCode string `json:"payment_code"`
	Amount      int64  `json:"amount"`
	QRImageURL  string `json:"qr_image_url"`
	ExpiresAt   string `json:"expires_at"`
	Message     string `json:"message"`
}
