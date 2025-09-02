package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"jinzmedia-atmt/database"
	"jinzmedia-atmt/models"
)

type DownloadService struct {
	userCollection     *mongo.Collection
	downloadCollection *mongo.Collection
}

func NewDownloadService() *DownloadService {
	return &DownloadService{
		userCollection:     database.GetCollection("users"),
		downloadCollection: database.GetCollection("downloads"),
	}
}

// GetProductsAndUserInfo returns available products and user information
func (ds *DownloadService) GetProductsAndUserInfo(userID primitive.ObjectID) (*models.ProductsResponse, error) {
	ctx := context.Background()

	// Get user information
	var user models.User
	err := ds.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create user info for response
	userInfo := models.UserInfo{
		Email:        user.Email,
		FullName:     user.FullName,
		Platform:     string(user.Platform),
		Owned:        user.Owned,
		SerialNumber: user.SerialNumber,
	}

	// Get products and check availability based on file existence
	products := make([]models.Product, 0, len(models.Products))
	for _, product := range models.Products {
		// Check if files exist for each platform
		availablePlatforms := make([]string, 0)
		for _, platform := range product.Platforms {
			filePath := ds.getProductFilePath(product.Name, platform)
			if _, err := os.Stat(filePath); err == nil {
				availablePlatforms = append(availablePlatforms, platform)
			}
		}
		
		productCopy := product
		productCopy.Platforms = availablePlatforms
		productCopy.Available = len(availablePlatforms) > 0
		products = append(products, productCopy)
	}

	response := &models.ProductsResponse{
		Products: products,
		User:     userInfo,
	}

	return response, nil
}

// ProcessDownloadRequest handles download validation and file serving
func (ds *DownloadService) ProcessDownloadRequest(userID primitive.ObjectID, productName, platform, serial string, r *http.Request) (*models.DownloadInfo, error) {
	ctx := context.Background()

	// Get user information
	var user models.User
	err := ds.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is banned
	if user.IsBanned {
		return nil, fmt.Errorf("user is banned")
	}

	// Check if user owns the product
	if !user.Owned {
		return nil, fmt.Errorf("you do not own this product")
	}

	// Validate serial number
	if user.SerialNumber != serial {
		return nil, fmt.Errorf("serial number does not match")
	}

	// Get file path
	filePath := ds.getProductFilePath(productName, platform)
	
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("failed to access file: %w", err)
	}

	// Log download
	err = ds.logDownload(userID, productName, platform, serial, r)
	if err != nil {
		// Don't fail the download if logging fails, just log the error
		fmt.Printf("Failed to log download: %v\n", err)
	}

	// Prepare download info
	filename := fmt.Sprintf("%s", filepath.Base(filePath))
	if platform == "windows" && filepath.Ext(filename) == "" {
		filename += ".exe"
	}

	return &models.DownloadInfo{
		FilePath: filePath,
		Filename: filename,
		Size:     fileInfo.Size(),
	}, nil
}

// getProductFilePath returns the file path for a product and platform
func (ds *DownloadService) getProductFilePath(productName, platform string) string {
	baseDir := "dist"
	return filepath.Join(baseDir, productName, platform, productName)
}

// logDownload records a download in the database
func (ds *DownloadService) logDownload(userID primitive.ObjectID, productName, platform, serial string, r *http.Request) error {
	ctx := context.Background()

	// Get client IP
	clientIP := r.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	// Get user agent
	userAgent := r.Header.Get("User-Agent")

	// Create download record
	downloadRecord := models.DownloadRecord{
		ID:           primitive.NewObjectID(),
		UserID:       userID,
		ProductName:  productName,
		Platform:     platform,
		SerialNumber: serial,
		IPAddress:    clientIP,
		UserAgent:    userAgent,
		DownloadedAt: time.Now(),
	}

	// Insert into database
	_, err := ds.downloadCollection.InsertOne(ctx, downloadRecord)
	if err != nil {
		return fmt.Errorf("failed to log download: %w", err)
	}

	return nil
}

// GetUserDownloadHistory returns download history for a user
func (ds *DownloadService) GetUserDownloadHistory(userID primitive.ObjectID) ([]*models.DownloadRecord, error) {
	ctx := context.Background()

	// Find downloads for the user
	cursor, err := ds.downloadCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("failed to find downloads: %w", err)
	}
	defer cursor.Close(ctx)

	var downloads []*models.DownloadRecord
	for cursor.Next(ctx) {
		var download models.DownloadRecord
		if err := cursor.Decode(&download); err != nil {
			return nil, fmt.Errorf("failed to decode download: %w", err)
		}
		downloads = append(downloads, &download)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return downloads, nil
}

// GetDownloadStats returns download statistics (for admin use)
func (ds *DownloadService) GetDownloadStats() (map[string]interface{}, error) {
	ctx := context.Background()

	// Count total downloads
	totalDownloads, err := ds.downloadCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count downloads: %w", err)
	}

	// Count downloads by product
	pipeline := []bson.M{
		{"$group": bson.M{
			"_id":   "$product_name",
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := ds.downloadCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate downloads: %w", err)
	}
	defer cursor.Close(ctx)

	productStats := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		productStats[result.ID] = result.Count
	}

	stats := map[string]interface{}{
		"total_downloads": totalDownloads,
		"by_product":      productStats,
	}

	return stats, nil
}
