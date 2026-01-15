package dtos

import (
	"time"

	"github.com/google/uuid"
)

type CreateMentorProfileRequest struct {
	Title    string `json:"title" binding:"required,min=3"`
	Bio      string `json:"bio"`
	Timezone string `json:"timezone" binding:"required"`
}

type CreateMentorProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title"`
	Bio       string    `json:"bio"`
	Timezone  string    `json:"timezone"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}
type MentorProfileResponse struct {
	User   MentorUserInfo `json:"user"`
	Mentor MentorInfo     `json:"mentor"`
}

type MentorUserInfo struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	ProfilePicture string    `json:"profile_picture"`
}

type MentorInfo struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Bio      string    `json:"bio"`
	Timezone string    `json:"timezone"`
	IsActive bool      `json:"is_active"`
}
