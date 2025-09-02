package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"jinzmedia-atmt/config"
)

var client *mongo.Client
var database *mongo.Database

// Connect establishes a connection to MongoDB
func Connect() error {
	cfg := config.Get()

	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.GetDatabaseDSN())

	// Set connection pool options
	clientOptions.SetMaxPoolSize(uint64(cfg.Database.MaxPoolSize))
	clientOptions.SetMinPoolSize(uint64(cfg.Database.MinPoolSize))
	clientOptions.SetConnectTimeout(cfg.Database.ConnectionTimeout)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectionTimeout)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Get database instance
	database = client.Database(cfg.Database.Name)

	log.Printf("Connected to MongoDB database: %s", cfg.Database.Name)
	return nil
}

// GetClient returns the MongoDB client
func GetClient() *mongo.Client {
	return client
}

// GetDatabase returns the MongoDB database
func GetDatabase() *mongo.Database {
	return database
}

// GetCollection returns a collection from the database
func GetCollection(name string) *mongo.Collection {
	if database == nil {
		panic("database not initialized. Call database.Connect() first")
	}
	return database.Collection(name)
}

// Disconnect closes the connection to MongoDB
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("Disconnected from MongoDB")
	return nil
}

// IsConnected checks if the database connection is alive
func IsConnected() bool {
	if client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Ping(ctx, nil)
	return err == nil
}
