// services/user-service/cmd/main.go
package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver

	"health-tracker-project/services/user-service/internal/handlers"
	"health-tracker-project/services/user-service/internal/repository"
	"health-tracker-project/services/user-service/internal/services"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the new logger package
)

func main() {
	// Initialize the logger first thing
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development" // Default to development environment
	}
	logger.InitLogger(env)
	defer logger.Logger.Sync() // Ensure all buffered logs are written when main exits

	logger.Logger.Info("Starting User Service...")

	// 1. Configuration (e.g., from environment variables)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Logger.Fatal("DATABASE_URL environment variable not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// 2. Initialize Repository (concrete implementation)
	// NewPostgresUserRepository handles DB connection, ping, and migrations internally.
	userRepo, err := repository.NewPostgresUserRepository(dbURL)
	if err != nil {
		logger.Logger.Fatalf("Failed to initialize user repository: %v", err)
	}
	// In a complete app, you might add a Close() method to UserRepository interface
	// and defer userRepo.Close() here for graceful shutdown of the DB connection.

	// 3. Initialize Service Implementations (concretions)
	// Services depend on repository interfaces.
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)

	// 4. Initialize Handler Implementations (concretions)
	// Handlers depend on service interfaces.
	authHandlers := handlers.NewAuthHandlers(authService)
	userHandlers := handlers.NewUserHandler(userService)

	// 5. Setup HTTP Router (using net/http's ServeMux with Go 1.22+ patterns)
	mux := http.NewServeMux()

	// Public Authentication Routes
	mux.HandleFunc("POST /register", authHandlers.Register)
	mux.HandleFunc("POST /login", authHandlers.Login)

	// Protected Authentication Routes (require JWT authentication middleware)
	mux.Handle("GET /protected", handlers.AuthMiddleware(http.HandlerFunc(authHandlers.ProtectedRoute)))
	mux.Handle("POST /logout", handlers.AuthMiddleware(http.HandlerFunc(authHandlers.Logout)))

	// User Management Routes (Protected)
	// Using the new Go 1.22+ pattern matching for path parameters
	mux.Handle("GET /users", handlers.AuthMiddleware(http.HandlerFunc(userHandlers.UsersCollectionHandler)))
	mux.Handle("POST /users", handlers.AuthMiddleware(http.HandlerFunc(userHandlers.UsersCollectionHandler)))
	mux.Handle("GET /users/{id}", handlers.AuthMiddleware(http.HandlerFunc(userHandlers.UserItemHandler)))
	mux.Handle("PUT /users/{id}", handlers.AuthMiddleware(http.HandlerFunc(userHandlers.UserItemHandler)))
	mux.Handle("DELETE /users/{id}", handlers.AuthMiddleware(http.HandlerFunc(userHandlers.UserItemHandler)))
	mux.Handle("GET /users/by-email", handlers.AuthMiddleware(http.HandlerFunc(userHandlers.GetUserByEmailHandler)))

	// Public Health Check Route
	mux.HandleFunc("GET /health", userHandlers.HealthCheck)

	// 6. Start HTTP Server
	logger.Logger.Infof("User Service listening on port %s", port)
	logger.Logger.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), mux))
}