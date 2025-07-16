package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"health-tracker-project/services/user-service/internal/handlers"
	"health-tracker-project/services/user-service/internal/repository"
)

func main() {
	// Database connection string from environment variable
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // Default for local Docker setup
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "usersdb"
	}

	dataSourceName := os.Getenv("DATABASE_URL")
	if dataSourceName == "" {
		dataSourceName = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPassword, dbName)
	}

	// Initialize User Repository
	userRepo, err := repository.NewPostgresUserRepository(dataSourceName)
	if err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}

	// Run migrations
	if err := userRepo.Migrate(); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Initialize Handler
	userHandler := handlers.NewUserHandler(userRepo)

	// Create a new ServeMux (router)
	mux := http.NewServeMux()

	// Register handlers with ServeMux, now with /v1 prefix
	// All user-related API endpoints will now start with /v1
	mux.HandleFunc("/v1/users", userHandler.UsersCollectionHandler)
	mux.HandleFunc("/v1/users/", userHandler.UserItemHandler) // /v1/users/ will match /v1/users/{id}
	mux.HandleFunc("/v1/users/by-email", userHandler.GetUserByEmailHandler)
	mux.HandleFunc("/health", userHandler.HealthCheck) // Health check typically not versioned, stays at root

	// Get port from environment, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("User Service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil { // Use the custom mux
		log.Fatalf("Could not start server: %v", err)
	}
}
