package dtos

import "github.com/google/uuid"

type UserProfileResponse struct {
	User     PublicUserResponse `json:"user"`
	IsMentor bool               `json:"is_mentor"`
	Mentor   *MentorPreview     `json:"mentor"`
}

type PublicUserResponse struct {
	ID             uuid.UUID `json:"id"`
	Username       string    `json:"username"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	ProfilePicture string    `json:"profile_picture"`
	Bio            string    `json:"bio"`
}

type MentorPreview struct {
	Title string `json:"title"`
}
