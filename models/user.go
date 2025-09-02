package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email" validate:"required,email"`
	Password     string             `bson:"password" json:"-"` // Never include password in JSON response
	FullName     string             `bson:"full_name" json:"full_name" validate:"required"`
	DateOfBirth  time.Time          `bson:"date_of_birth" json:"date_of_birth"`
	Platform     Platform           `bson:"platform" json:"platform"`
	Owned        bool               `bson:"owned" json:"owned"`
	IsBanned     bool               `bson:"is_banned" json:"is_banned"`
	SerialNumber string             `bson:"serial_number" json:"serial_number"`
	PaymentCode  string             `bson:"payment_code,omitempty" json:"payment_code,omitempty"`
	Role         string             `bson:"role" json:"role"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	LastLogin    *time.Time         `bson:"last_login,omitempty" json:"last_login,omitempty"`
}

// Platform represents supported platforms
type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformMacOS   Platform = "macos"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
	RoleSuper UserRole = "super"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email        string    `json:"email" validate:"required,email"`
	Password     string    `json:"password" validate:"required,min=6"`
	FullName     string    `json:"full_name" validate:"required"`
	DateOfBirth  time.Time `json:"date_of_birth"`
	Platform     Platform  `json:"platform" validate:"required"`
	SerialNumber string    `json:"serial_number" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	User         *User  `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	Type   string   `json:"type"` // "access" or "refresh"
}
