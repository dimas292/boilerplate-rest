package auth

import (
	"github.com/dimas292/url_shortener/pkg/model"
)

// User is the domain model for authentication.
type User struct {
	model.BaseModel
	Name     string `json:"name" gorm:"type:varchar(100);not null" binding:"required"`
	Email    string `json:"email" gorm:"type:varchar(255);uniqueIndex;not null" binding:"required,email"`
	Password string `json:"-" gorm:"type:varchar(255);not null"` // json:"-" hides password from responses
	Role     string `json:"role" gorm:"type:varchar(20);default:'user'"`
}

// RegisterRequest is the DTO for user registration.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest is the DTO for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is the response containing user info and JWT token.
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// UserResponse is the safe user representation (no password).
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// ToResponse converts a User to a safe UserResponse.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Role:  u.Role,
	}
}
