// services/user-service/internal/services/interfaces.go
package services

import (
	"github.com/google/uuid"
	"health-tracker-project/services/user-service/internal/models"
)

// AuthService defines the interface for authentication-related business logic.
type AuthService interface {
	RegisterUser(req models.RegisterRequest) (*models.UserResponse, error)
	AuthenticateUser(req models.LoginRequest) (*models.AuthResponse, error)
	// Add other authentication-related methods if needed, e.g., ResetPassword, VerifyEmail
}

// UserService defines the interface for general user-related business logic.
type UserService interface {
	CreateUser(req models.CreateUserRequest) (*models.UserResponse, error)
	GetUserByID(id uuid.UUID) (*models.UserResponse, error)
	GetAllUsers() ([]models.UserResponse, error)
	GetUserByEmail(email string) (*models.UserResponse, error)
	UpdateUser(id uuid.UUID, req models.UpdateUserRequest) (*models.UserResponse, error)
	DeleteUser(id uuid.UUID) error
}