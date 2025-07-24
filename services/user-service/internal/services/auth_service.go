// services/user-service/internal/services/auth_service.go
package services

import (
	"fmt"
	"time"

	"health-tracker-project/services/user-service/internal/models"
	"health-tracker-project/services/user-service/internal/repository"
	"health-tracker-project/services/user-service/internal/utils/jwt"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the logger
)

// AuthServiceImpl implements the AuthService interface.
type AuthServiceImpl struct {
	userRepo repository.UserRepository // Depends on the UserRepository interface
}

// NewAuthService creates a new instance of AuthServiceImpl.
func NewAuthService(userRepo repository.UserRepository) *AuthServiceImpl {
	return &AuthServiceImpl{userRepo: userRepo}
}

// RegisterUser handles the business logic for new user registration.
func (s *AuthServiceImpl) RegisterUser(req models.RegisterRequest) (*models.UserResponse, error) {
	// Business validation: Ensure all required fields are present.
	if req.Name == "" || req.Email == "" || req.Password == "" {
		logger.Logger.Debug("Registration request missing required fields.")
		return nil, fmt.Errorf("service: name, email, and password are required")
	}
	// Add more robust validation here (e.g., email format, password strength).

	// Check if user with this email already exists.
	existingUser, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Failed to check for existing user by email '%s': %v", req.Email, err)
		return nil, fmt.Errorf("service: failed to check for existing user by email: %w", err)
	}
	if existingUser != nil {
		logger.Logger.Warnf("Registration attempt with existing email: %s", req.Email)
		return nil, fmt.Errorf("service: user with this email already exists")
	}

	// Create new user model (password hashing is handled inside models.NewUser).
	newUser, err := models.NewUser(req.Name, req.Email, req.Password)
	if err != nil {
		logger.Logger.Errorf("Failed to create new user model: %v", err)
		return nil, fmt.Errorf("service: failed to create new user model: %w", err)
	}

	// Persist the user to the database via the repository.
	if err := s.userRepo.CreateUser(newUser); err != nil {
		logger.Logger.Errorf("Failed to save new user '%s': %v", newUser.ID, err)
		return nil, fmt.Errorf("service: failed to save new user: %w", err)
	}

	userResponse := newUser.ToUserResponse()
	logger.Logger.Infof("User registered successfully: ID %s, Email %s", newUser.ID, newUser.Email)
	return &userResponse, nil
}

// AuthenticateUser handles the business logic for user login.
func (s *AuthServiceImpl) AuthenticateUser(req models.LoginRequest) (*models.AuthResponse, error) {
	// Business validation: Ensure required fields for login are present.
	if req.Email == "" || req.Password == "" {
		logger.Logger.Debug("Login request missing email or password.")
		return nil, fmt.Errorf("service: email and password are required")
	}

	// Retrieve user by email from the repository.
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Failed to retrieve user by email '%s' for authentication: %v", req.Email, err)
		return nil, fmt.Errorf("service: failed to retrieve user for authentication: %w", err)
	}
	// Check if user exists and if password is correct.
	if user == nil || !user.CheckPassword(req.Password) {
		logger.Logger.Warnf("Invalid login attempt for email '%s'.", req.Email)
		return nil, fmt.Errorf("service: invalid credentials")
	}

	// Generate JWT upon successful authentication.
	tokenDuration := 15 * time.Minute // Short-lived access token
	// Generate JWT using user's ID and Name for claims.
	tokenString, err := jwt.GenerateJWT(user.ID.String(), user.Name, tokenDuration)
	if err != nil {
		logger.Logger.Errorf("Failed to generate JWT for user '%s': %v", user.ID, err)
		return nil, fmt.Errorf("service: failed to generate token: %w", err)
	}

	logger.Logger.Infof("User authenticated successfully: ID %s, Email %s", user.ID, user.Email)
	return &models.AuthResponse{
		Token:        tokenString,
		User:         user.ToUserResponse(),
		ExpiresInSec: int64(tokenDuration.Seconds()),
	}, nil
}