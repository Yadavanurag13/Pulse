// services/user-service/internal/services/user_service.go
package services

import (
	"fmt"

	"github.com/google/uuid"
	"health-tracker-project/services/user-service/internal/models"
	"health-tracker-project/services/user-service/internal/repository"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the logger
)

// UserServiceImpl implements the UserService interface.
type UserServiceImpl struct {
	userRepo repository.UserRepository // Depends on the UserRepository interface
}

// NewUserService creates a new instance of UserServiceImpl.
func NewUserService(userRepo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{userRepo: userRepo}
}

// CreateUser handles the business logic for creating a new user (e.g., by an admin).
func (s *UserServiceImpl) CreateUser(req models.CreateUserRequest) (*models.UserResponse, error) {
	// Business validation
	if req.Name == "" || req.Email == "" || req.Password == "" {
		logger.Logger.Debug("CreateUser request missing required fields.")
		return nil, fmt.Errorf("service: name, email, and password are required")
	}

	// Check if user with this email already exists
	existingUser, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		logger.Logger.Errorf("Failed to check for existing user by email '%s': %v", req.Email, err)
		return nil, fmt.Errorf("service: failed to check for existing user by email: %w", err)
	}
	if existingUser != nil {
		logger.Logger.Warnf("CreateUser attempt with existing email: %s", req.Email)
		return nil, fmt.Errorf("service: user with this email already exists")
	}

	// Create new user model (password hashing handled inside NewUser)
	newUser, err := models.NewUser(req.Name, req.Email, req.Password)
	if err != nil {
		logger.Logger.Errorf("Failed to create new user model: %v", err)
		return nil, fmt.Errorf("service: failed to create new user model: %w", err)
	}

	// Persist user to database
	if err := s.userRepo.CreateUser(newUser); err != nil {
		logger.Logger.Errorf("Failed to save new user '%s': %v", newUser.ID, err)
		return nil, fmt.Errorf("service: failed to save new user: %w", err)
	}

	userResponse := newUser.ToUserResponse()
	logger.Logger.Infof("User created via admin/service: ID %s, Email %s", newUser.ID, newUser.Email)
	return &userResponse, nil
}

// GetUserByID retrieves a user by their ID.
func (s *UserServiceImpl) GetUserByID(id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		logger.Logger.Errorf("Failed to retrieve user by ID '%s': %v", id, err)
		return nil, fmt.Errorf("service: failed to retrieve user by ID: %w", err)
	}
	if user == nil {
		logger.Logger.Debugf("User with ID '%s' not found.", id)
		return nil, fmt.Errorf("service: user not found")
	}
	userResponse := user.ToUserResponse()
	logger.Logger.Debugf("Retrieved user by ID: %s", id)
	return &userResponse, nil
}

// GetAllUsers retrieves all users.
func (s *UserServiceImpl) GetAllUsers() ([]models.UserResponse, error) {
	users, err := s.userRepo.GetAllUsers()
	if err != nil {
		logger.Logger.Errorf("Failed to retrieve all users: %v", err)
		return nil, fmt.Errorf("service: failed to retrieve all users: %w", err)
	}

	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToUserResponse()
	}
	logger.Logger.Debugf("Retrieved %d users.", len(userResponses))
	return userResponses, nil
}

// GetUserByEmail retrieves a user by their email address.
func (s *UserServiceImpl) GetUserByEmail(email string) (*models.UserResponse, error) {
	if email == "" {
		logger.Logger.Debug("GetUserByEmail request missing email.")
		return nil, fmt.Errorf("service: email is required")
	}

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		logger.Logger.Errorf("Failed to retrieve user by email '%s': %v", email, err)
		return nil, fmt.Errorf("service: failed to retrieve user by email: %w", err)
	}
	if user == nil {
		logger.Logger.Debugf("User with email '%s' not found.", email)
		return nil, fmt.Errorf("service: user not found")
	}
	userResponse := user.ToUserResponse()
	logger.Logger.Debugf("Retrieved user by email: %s", email)
	return &userResponse, nil
}

// UpdateUser updates an existing user's details.
func (s *UserServiceImpl) UpdateUser(id uuid.UUID, req models.UpdateUserRequest) (*models.UserResponse, error) {
	// Retrieve existing user
	existingUser, err := s.userRepo.GetUserByID(id)
	if err != nil {
		logger.Logger.Errorf("Failed to retrieve user '%s' for update: %v", id, err)
		return nil, fmt.Errorf("service: failed to retrieve user for update: %w", err)
	}
	if existingUser == nil {
		logger.Logger.Warnf("User '%s' not found for update.", id)
		return nil, fmt.Errorf("service: user not found for update")
	}

	// Apply updates based on provided fields in the request
	if req.Name != "" {
		existingUser.Name = req.Name
	}
	if req.Email != "" {
		// If email is changed, check for uniqueness among other users
		if req.Email != existingUser.Email {
			userWithNewEmail, err := s.userRepo.GetUserByEmail(req.Email)
			if err != nil {
				logger.Logger.Errorf("Failed to check for email uniqueness for user '%s' with new email '%s': %v", id, req.Email, err)
				return nil, fmt.Errorf("service: failed to check for email uniqueness: %w", err)
			}
			if userWithNewEmail != nil && userWithNewEmail.ID != existingUser.ID {
				logger.Logger.Warnf("Update for user '%s' failed, new email '%s' already in use.", id, req.Email)
				return nil, fmt.Errorf("service: new email already in use by another user")
			}
		}
		existingUser.Email = req.Email
	}
	if req.Password != nil && *req.Password != "" { // Check if password is provided and not empty
		// Use models.NewUser to hash the new password.
		// We create a temporary user just for its password hashing capability.
		tempUserWithHashedPwd, err := models.NewUser("", "", *req.Password)
		if err != nil {
			logger.Logger.Errorf("Failed to hash new password for user '%s': %v", id, err)
			return nil, fmt.Errorf("service: failed to hash new password: %w", err)
		}
		existingUser.PasswordHash = tempUserWithHashedPwd.PasswordHash
	}

	// Persist updated user
	if err := s.userRepo.UpdateUser(existingUser); err != nil {
		logger.Logger.Errorf("Failed to update user '%s': %v", id, err)
		return nil, fmt.Errorf("service: failed to update user: %w", err)
	}

	userResponse := existingUser.ToUserResponse()
	logger.Logger.Infof("User updated: %s", userResponse.ID)
	return &userResponse, nil
}

// DeleteUser deletes a user by their ID.
func (s *UserServiceImpl) DeleteUser(id uuid.UUID) error {
	// Optional: Check if user exists before attempting delete to return a more specific "not found" error.
	// This adds a DB lookup but provides clearer API responses.
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		logger.Logger.Errorf("Failed to check user existence before deleting user '%s': %v", id, err)
		return fmt.Errorf("service: failed to check user existence before delete: %w", err)
	}
	if user == nil {
		logger.Logger.Warnf("Deletion failed, user '%s' not found.", id)
		return fmt.Errorf("service: user not found for deletion")
	}

	if err := s.userRepo.DeleteUser(id); err != nil {
		logger.Logger.Errorf("Failed to delete user '%s': %v", id, err)
		return fmt.Errorf("service: failed to delete user: %w", err)
	}
	logger.Logger.Infof("User deleted: %s", id)
	return nil
}