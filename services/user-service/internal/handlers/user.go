package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings" // Used for parsing URL paths

	"golang.org/x/crypto/bcrypt"
	"health-tracker-project/services/user-service/internal/models"
	"health-tracker-project/services/user-service/internal/repository"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// UsersCollectionHandler handles /users for GET (all users) and POST (create user)
func (h *UserHandler) UsersCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAllUsers(w, r)
	case http.MethodPost:
		h.CreateUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// UserItemHandler handles /users/{id} for GET, PUT, DELETE
func (h *UserHandler) UserItemHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from the URL path.
	// r.URL.Path could be "/v1/users/some-id"
	// We need to get the part after the last slash.
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 2 { // Should at least have "", "v1", "users", "ID"
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	id := parts[len(parts)-1] // Get the last part of the path

	if id == "" { // Basic validation for ID presence
		http.Error(w, "User ID is required in path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetUserByID(w, r, id)
	case http.MethodPut:
		h.UpdateUser(w, r, id)
	case http.MethodDelete:
		h.DeleteUser(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetUserByEmailHandler handles /users/by-email?email=...
func (h *UserHandler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	 if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
	h.GetUserByEmail(w, r)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = h.userRepo.CreateUser(user)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			http.Error(w, "User with this email already exists", http.StatusConflict)
			return
		}
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userResp := models.ToUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(userResp)
	log.Printf("User created: %s", user.ID)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request, id string) {
	// Implementation for getting a user by ID
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user by ID: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	userResp := models.ToUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
	log.Printf("User retrieved: %s", user.ID)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.userRepo.GetAllUsers()
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}
	
	usersResp := make([]models.UserResponse, len(users))
	for i, user := range users {
		usersResp[i] = models.ToUserResponse(&user)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usersResp)
	log.Printf("Retrieved %d users", len(users))
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email query parameter is required", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetUserByEmail(email)
	if err != nil {
		if strings.Contains(err.Error(), "user with email not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Error getting user by email: %v", err)
		http.Error(w, "Failed to get user", http.StatusInternalServerError)
		return
	}
	userResp := models.ToUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
	log.Printf("User retrieved by email: %s", user.Email)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}


	if req.Name == "" || req.Email == "" || req.Password == "" {
        http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
        return
    }

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user := &models.User{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = h.userRepo.UpdateUser(user)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	userResp := models.ToUserResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
	log.Printf("User updated: %s", user.ID)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if id == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	err := h.userRepo.DeleteUser(id)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("User deleted: %s", id)
}

func (h *UserHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("User Service is healthy"))
}