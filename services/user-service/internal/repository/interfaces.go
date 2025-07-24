// services/user-service/internal/repository/interfaces.go
package repository

import (
	"github.com/google/uuid"
	"health-tracker-project/services/user-service/internal/models"
)

// UserRepository defines the interface for user data operations.
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uuid.UUID) error
	Migrate() error // Method to run database migrations
}