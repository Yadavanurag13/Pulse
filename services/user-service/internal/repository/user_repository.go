// services/user-service/internal/repository/user_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // PostgreSQL driver

	"health-tracker-project/services/user-service/internal/models"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the logger
)

// postgresUserRepository is the concrete implementation of UserRepository for PostgreSQL.
type postgresUserRepository struct {
	db *sql.DB // The standard Go SQL database connection pool
}

// NewPostgresUserRepository creates a new instance of PostgresUserRepository,
// connects to the database, pings it, and runs migrations.
// It returns the UserRepository interface, adhering to Dependency Inversion Principle.
func NewPostgresUserRepository(dataSourceName string) (UserRepository, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Ping the database to ensure connection is established
	if err = db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &postgresUserRepository{db: db}

	// Run migrations (e.g., create tables if they don't exist)
	if err := repo.Migrate(); err != nil {
		db.Close() // Close the connection if migration fails
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	logger.Logger.Info("Connected to PostgreSQL database successfully!")
	return repo, nil
}

// Migrate creates the 'users' table if it doesn't exist.
// Column names use standard snake_case for PostgreSQL.
func (r *postgresUserRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL, -- Email is unique and used for login
		password_hash VARCHAR(255) NOT NULL, -- Storing the bcrypt hashed password
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	logger.Logger.Info("Database migration completed successfully!")
	return nil
}

// CreateUser inserts a new user into the database.
// It assumes the user ID and timestamps are set by the models.NewUser constructor.
func (r *postgresUserRepository) CreateUser(user *models.User) error {
	// Defensive check, user.ID should be set by models.NewUser
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	// Ensure timestamps are UTC for consistency
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = user.CreatedAt

	query := `INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("repository: failed to create user: %w", err)
	}
	logger.Logger.Infof("User created successfully: %s", user.ID)
	return nil
}

// GetUserByEmail retrieves a user by their email address.
// This is intended to be the primary lookup for authentication.
func (r *postgresUserRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`
	row := r.db.QueryRow(query, email)

	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			logger.Logger.Debugf("User with email '%s' not found in DB.", email)
			return nil, nil // Return nil, nil when user is not found (idiomatic Go)
		}
		return nil, fmt.Errorf("repository: failed to get user by email: %w", err)
	}
	logger.Logger.Debugf("Retrieved user by email '%s': %s", email, user.ID)
	return &user, nil
}

// GetAllUsers retrieves all users from the database.
func (r *postgresUserRepository) GetAllUsers() ([]models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("repository: failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository: failed to scan user row: %w", err)
		}
		users = append(users, user)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows iteration error: %w", err)
	}
	logger.Logger.Debugf("Retrieved %d users from DB.", len(users))
	return users, nil
}

// GetUserByID retrieves a user by their UUID.
func (r *postgresUserRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE id = $1`
	row := r.db.QueryRow(query, id)

	var user models.User
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			logger.Logger.Debugf("User with ID '%s' not found in DB.", id)
			return nil, nil // Return nil, nil when user is not found
		}
		return nil, fmt.Errorf("repository: failed to get user by ID: %w", err)
	}
	logger.Logger.Debugf("Retrieved user by ID '%s': %s", id, user.Name)
	return &user, nil
}

// UpdateUser updates an existing user's details in the database.
func (r *postgresUserRepository) UpdateUser(user *models.User) error {
	user.UpdatedAt = time.Now().UTC() // Update timestamp on modification

	query := `UPDATE users SET name = $1, email = $2, password_hash = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.Exec(query, user.Name, user.Email, user.PasswordHash, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("repository: failed to update user: %w", err)
	}
	logger.Logger.Infof("User updated successfully: %s", user.ID)
	return nil
}

// DeleteUser deletes a user from the database by their UUID.
func (r *postgresUserRepository) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("repository: failed to delete user: %w", err)
	}
	logger.Logger.Infof("User deleted successfully: %s", id)
	return nil
}