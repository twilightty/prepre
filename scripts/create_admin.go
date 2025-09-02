package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"jinzmedia-atmt/auth"
	"jinzmedia-atmt/config"
	"jinzmedia-atmt/database"
	"jinzmedia-atmt/models"
)

func main() {
	if len(os.Args) < 6 {
		fmt.Println("Usage: go run scripts/create_admin.go <email> <password> <first_name> <last_name> <role>")
		fmt.Println("Roles: user, admin, super")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]
	firstName := os.Args[3]
	lastName := os.Args[4]
	roleStr := os.Args[5]

	// Validate role
	var role models.UserRole
	switch roleStr {
	case "user":
		role = models.RoleUser
	case "admin":
		role = models.RoleAdmin
	case "super":
		role = models.RoleSuper
	default:
		log.Fatalf("Invalid role: %s. Must be one of: user, admin, super", roleStr)
	}

	// Load configuration
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Disconnect()

	// Create auth service
	authService := auth.NewAuthService()

	// Create user
	req := &models.RegisterRequest{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
	}

	user, err := authService.Register(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	// Update user role if not default
	if role != models.RoleUser {
		collection := database.GetCollection("users")
		_, err = collection.UpdateOne(
			context.Background(),
			map[string]interface{}{"_id": user.ID},
			map[string]interface{}{"$set": map[string]interface{}{"role": role}},
		)
		if err != nil {
			log.Fatalf("Failed to update user role: %v", err)
		}
	}

	fmt.Printf("Successfully created %s user: %s (%s %s)\n", role, email, firstName, lastName)
}
