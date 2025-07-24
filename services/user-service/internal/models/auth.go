// services/user-service/internal/models/auth.go
package models

// LoginRequest defines the structure for a login request from the client.
// It uses 'email' as the primary identifier for consistency with GetUserByEmail.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest defines the structure for a user registration request from the client.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse defines the structure for a successful authentication response to the client.
type AuthResponse struct {
	Token        string       `json:"token"`
	User         UserResponse `json:"user"` // Uses the UserResponse DTO from models/user.go
	ExpiresInSec int64        `json:"expires_in_sec"`
}