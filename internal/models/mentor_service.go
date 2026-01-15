package models

import (
	"time"

	"github.com/google/uuid"
)

type MentorService struct {
	ID              uuid.UUID
	MentorID        uuid.UUID
	Title           string
	Description     string
	DurationMinutes int
	PriceCents      int
	Currency        string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
