package models

import "time"

type User struct {
	ID        string    `json:"id,omitempty"` 
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"` 
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ErrorResponse for API error messages
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}