// services/user-service/internal/handlers/auth.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"health-tracker-project/services/user-service/internal/models"
	"health-tracker-project/services/user-service/internal/services"
	"health-tracker-project/services/user-service/internal/utils/jwt"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the logger
)

// ContextKey type for storing values in request context.
type ContextKey string

const UserContextKey ContextKey = "user" // Key to store user ID in context

// AuthHandlers holds dependencies for authentication HTTP handlers.
type AuthHandlers struct {
	authService services.AuthService // Depends on the AuthService interface
}

// NewAuthHandlers creates a new AuthHandlers instance.
func NewAuthHandlers(authService services.AuthService) *AuthHandlers {
	return &AuthHandlers{authService: authService}
}

// Register handles HTTP requests for new user registration.
func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Debugf("Invalid request payload for register: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userResponse, err := h.authService.RegisterUser(req) // Call the service layer
	if err != nil {
		// Map service-level errors to appropriate HTTP status codes
		if err.Error() == "service: user with this email already exists" {
			logger.Logger.Warnf("Registration failed: %v", err)
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
		} else if err.Error() == "service: name, email, and password are required" {
			logger.Logger.Warnf("Registration failed: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest) // 400 Bad Request
		} else {
			logger.Logger.Errorf("Error registering user: %v", err)
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResponse)
	logger.Logger.Infof("User registered successfully: %s", userResponse.ID)
}

// Login handles HTTP requests for user login.
func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Debugf("Invalid request payload for login: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	authResponse, err := h.authService.AuthenticateUser(req) // Call the service layer
	if err != nil {
		if err.Error() == "service: invalid credentials" {
			logger.Logger.Warnf("Authentication failed for email '%s': %v", req.Email, err)
			http.Error(w, err.Error(), http.StatusUnauthorized) // 401 Unauthorized
		} else if err.Error() == "service: email and password are required" {
			logger.Logger.Warnf("Authentication failed (missing fields): %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest) // 400 Bad Request
		} else {
			logger.Logger.Errorf("Error during login for email '%s': %v", req.Email, err)
			http.Error(w, "Failed to authenticate", http.StatusInternalServerError)
		}
		return
	}

	// Set HttpOnly cookie for the JWT token
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    authResponse.Token,
		Expires:  time.Now().Add(time.Duration(authResponse.ExpiresInSec) * time.Second),
		HttpOnly: true,                 // Crucial for security (prevents JS access)
		Secure:   false,                // Set to 'true' in production with HTTPS (e.g., in a Dockerfile or deployment config)
		SameSite: http.SameSiteLaxMode, // Adjust as needed (Strict, Lax, None). Use http.SameSiteNone and Secure:true for cross-origin if frontend is on different domain/port.
		Path:     "/",                  // Available to all paths
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResponse)
	logger.Logger.Infof("User logged in successfully: %s", authResponse.User.ID)
}

// Logout handles HTTP requests for user logout by clearing the JWT cookie.
func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	// Invalidate the JWT cookie by setting an expired cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    "",
		Expires:  time.Unix(0, 0), // Set expiry to past
		HttpOnly: true,
		Secure:   false, // Set to 'true' in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
	logger.Logger.Info("User logged out successfully.")
}

// ProtectedRoute is an example handler that demonstrates JWT authentication.
func (h *AuthHandlers) ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	// User ID is extracted from the JWT and placed in the request context by AuthMiddleware.
	userID, ok := r.Context().Value(UserContextKey).(string)
	if !ok {
		// This case should ideally not be reached if AuthMiddleware is correctly applied
		logger.Logger.Error("User ID not found in context for protected route, middleware error?")
		http.Error(w, "Internal server error: User ID not found in context", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": fmt.Sprintf("Welcome to the protected area, User ID: %s!", userID)})
	logger.Logger.Debugf("Accessed protected route by User ID: %s", userID)
}

// AuthMiddleware is an HTTP middleware for JWT authentication.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt_token")
		if err != nil {
			if err == http.ErrNoCookie {
				logger.Logger.Debug("Unauthorized: No JWT token cookie found.")
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}
			logger.Logger.Warnf("Bad request: error reading JWT cookie: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest) // Malformed cookie header
			return
		}

		tokenString := cookie.Value
		claims, err := jwt.ParseJWT(tokenString) // Validate token using JWT utility
		if err != nil {
			logger.Logger.Warnf("Unauthorized: Invalid JWT token: %v", err)
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID (from JWT claims) to the request context for downstream handlers.
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserContextKey, claims.UserID)
		r = r.WithContext(ctx)

		logger.Logger.Debugf("JWT authentication successful for User ID: %s", claims.UserID)
		next.ServeHTTP(w, r)
	})
}