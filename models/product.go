package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a downloadable product
type Product struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Available   bool     `json:"available"`
	Platforms   []string `json:"platforms"`
}

// Available products
var Products = []Product{
	{Name: "chatgpt", DisplayName: "ChatGPT", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "dalle", DisplayName: "DALL-E", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "gemini", DisplayName: "Gemini", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "hailuo", DisplayName: "Hailuo", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "runway", DisplayName: "Runway", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "sora", DisplayName: "Sora", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "veo3", DisplayName: "Veo 3", Available: true, Platforms: []string{"windows", "macos"}},
	{Name: "veo3_pro", DisplayName: "Veo 3 Pro", Available: true, Platforms: []string{"windows", "macos"}},
}

// ProductsResponse represents the response for listing products
type ProductsResponse struct {
	Products []Product `json:"products"`
	User     UserInfo  `json:"user"`
}

// UserInfo represents user information in product response
type UserInfo struct {
	Email        string `json:"email"`
	FullName     string `json:"full_name"`
	Platform     string `json:"platform"`
	Owned        bool   `json:"owned"`
	SerialNumber string `json:"serial_number"`
}

// DownloadRecord represents a download history record
type DownloadRecord struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	ProductName  string             `bson:"product_name" json:"product_name"`
	Platform     string             `bson:"platform" json:"platform"`
	SerialNumber string             `bson:"serial_number" json:"serial_number"`
	IPAddress    string             `bson:"ip_address" json:"ip_address"`
	UserAgent    string             `bson:"user_agent" json:"user_agent"`
	DownloadedAt time.Time          `bson:"downloaded_at" json:"downloaded_at"`
}

// DownloadInfo represents download file information
type DownloadInfo struct {
	FilePath string
	Filename string
	Size     int64
}

// IsValidProduct checks if a product name is valid
func IsValidProduct(productName string) bool {
	for _, product := range Products {
		if product.Name == productName {
			return true
		}
	}
	return false
}

// GetProduct returns product by name
func GetProduct(productName string) (Product, bool) {
	for _, product := range Products {
		if product.Name == productName {
			return product, true
		}
	}
	return Product{}, false
}

// GetProductPlatforms returns available platforms for a product
func GetProductPlatforms(productName string) []string {
	product, exists := GetProduct(productName)
	if !exists {
		return []string{}
	}
	return product.Platforms
}

// IsValidPlatform checks if a platform is valid for a product
func IsValidPlatform(productName, platform string) bool {
	platforms := GetProductPlatforms(productName)
	for _, p := range platforms {
		if p == platform {
			return true
		}
	}
	return false
}
