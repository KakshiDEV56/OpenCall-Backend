package dtos

import "github.com/google/uuid"

type CreateMentorServiceRequest struct {
	Title           string `json:"title" binding:"required,min=3"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes" binding:"required,oneof=30 60"`
	PriceCents      int    `json:"price_cents" binding:"required,min=0"`
	Currency        string `json:"currency" binding:"required,len=3"`
}

type MentorServiceResponse struct {
	ID              uuid.UUID `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	PriceCents      int       `json:"price_cents"`
	Currency        string    `json:"currency"`
	IsActive        bool      `json:"is_active"`
}
