package dtos

import (
	"time"

	"github.com/google/uuid"
)

// Sending Response to the client of the user
type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	ProfilePicture string    `json:"profile_picture"`
	Bio            string    `json:"bio"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

// client will send request to register a user in this format
type RegisterUserRequest struct {
	FirstName string `json:"first_name" binding:"required,min=2,max=50"`
	LastName  string `json:"last_name" binding:"required,min=2,max=50"`
	Username  string `json:"username" binding:"required,min=3,max=30"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
}

// Sending response to the client after registering a user
type RegisterUserResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}

// client will send request to login a user in this format
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // can be username or email
	Password   string `json:"password" binding:"required"`
}

// Sending response to the client after logging in a user
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
}

// client will send request to refresh access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Sending response to the client after refreshing access token
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// client will send request to update user details
type UpdateUserRequest struct {
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	ProfilePicture *string `json:"profile_picture,omitempty"`
}

// Sending response to the client after updating user details
type UpdateUserResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}
