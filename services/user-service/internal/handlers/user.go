// services/user-service/internal/handlers/user.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"health-tracker-project/services/user-service/internal/models"
	"health-tracker-project/services/user-service/internal/services"
	"health-tracker-project/services/user-service/internal/utils/logger" // Import the logger
)

// UserHandler holds dependencies for user-related HTTP handlers.
type UserHandler struct {
	userService services.UserService // Depends on the UserService interface
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// UsersCollectionHandler routes requests to /users (GET all, POST create).
func (h *UserHandler) UsersCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAllUsers(w, r)
	case http.MethodPost:
		h.CreateUser(w, r)
	default:
		logger.Logger.Warnf("Method not allowed for /users: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserItemHandler routes requests to /users/{id} (GET, PUT, DELETE).
func (h *UserHandler) UserItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL path using Go 1.22+ PathValue or manual splitting
	idParam := r.PathValue("id")
	if idParam == "" {
		// Fallback for older Go or if PathValue isn't used
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			idParam = parts[len(parts)-1]
		}
	}

	if idParam == "" {
		logger.Logger.Debug("User ID is missing from path for item handler.")
		http.Error(w, "User ID is required in path", http.StatusBadRequest)
		return
	}

	// Convert string ID from URL to uuid.UUID for service layer
	userID, err := uuid.Parse(idParam)
	if err != nil {
		logger.Logger.Warnf("Invalid user ID format '%s': %v", idParam, err)
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetUserByID(w, r, userID)
	case http.MethodPut:
		h.UpdateUser(w, r, userID)
	case http.MethodDelete:
		h.DeleteUser(w, r, userID)
	default:
		logger.Logger.Warnf("Method not allowed for /users/{id}: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetUserByEmailHandler routes GET requests to /users/by-email?email=...
func (h *UserHandler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Logger.Warnf("Method not allowed for /users/by-email: %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.GetUserByEmail(w, r)
}

// CreateUser handles POST /users requests to create a new user.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Debugf("Invalid request payload for create user: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userResp, err := h.userService.CreateUser(req) // Call the service layer
	if err != nil {
		// Map service-level errors to HTTP status codes (simplified with string checks)
		if strings.Contains(err.Error(), "already exists") {
			logger.Logger.Warnf("User creation failed (conflict): %v", err)
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
		} else if strings.Contains(err.Error(), "required") {
			logger.Logger.Warnf("User creation failed (missing fields): %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest) // 400 Bad Request
		} else {
			logger.Logger.Errorf("Error creating user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResp)
	logger.Logger.Infof("User created: %s", userResp.ID)
}

// GetUserByID handles GET /users/{id} requests to retrieve a user by ID.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	userResp, err := h.userService.GetUserByID(id) // Call the service layer
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Logger.Warnf("User not found by ID: %s", id)
			http.Error(w, err.Error(), http.StatusNotFound) // 404 Not Found
		} else {
			logger.Logger.Errorf("Error getting user by ID %s: %v", id, err)
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
	logger.Logger.Infof("User retrieved by ID: %s", userResp.ID)
}

// GetAllUsers handles GET /users requests to retrieve all users.
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	usersResp, err := h.userService.GetAllUsers() // Call the service layer
	if err != nil {
		logger.Logger.Errorf("Error getting all users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usersResp)
	logger.Logger.Infof("Retrieved %d users", len(usersResp))
}

// GetUserByEmail handles GET /users/by-email?email=... requests.
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		logger.Logger.Debug("Email query parameter is missing for GetUserByEmail.")
		http.Error(w, "Email query parameter is required", http.StatusBadRequest)
		return
	}

	userResp, err := h.userService.GetUserByEmail(email) // Call the service layer
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Logger.Warnf("User not found by email: %s", email)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "required") {
			logger.Logger.Warnf("User retrieval by email failed (missing fields): %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			logger.Logger.Errorf("Error getting user by email %s: %v", email, err)
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
	logger.Logger.Infof("User retrieved by email: %s", userResp.Email)
}

// UpdateUser handles PUT /users/{id} requests to update user details.
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Debugf("Invalid request payload for update user %s: %v", id, err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userResp, err := h.userService.UpdateUser(id, req) // Call the service layer
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			logger.Logger.Warnf("User not found for update: %s", id)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "already in use") || strings.Contains(err.Error(), "required") {
			logger.Logger.Warnf("User update failed (validation/conflict): %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			logger.Logger.Errorf("Error updating user %s: %v", id, err)
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
	logger.Logger.Infof("User updated: %s", userResp.ID)
}

// DeleteUser handles DELETE /users/{id} requests to delete a user.
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id uuid.UUID) {
	err := h.userService.DeleteUser(id) // Call the service layer
	if err != nil {
		if strings.Contains(err.Error(), "not found") { // If service checks for existence
			logger.Logger.Warnf("User not found for deletion: %s", id)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			logger.Logger.Errorf("Error deleting user %s: %v", id, err)
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
	logger.Logger.Infof("User deleted: %s", id)
}

// HealthCheck provides a simple health check endpoint.
func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("User Service is healthy"))
	logger.Logger.Debug("Health check requested and passed.")
}