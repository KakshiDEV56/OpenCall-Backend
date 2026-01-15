package dtos

import (
	"github.com/google/uuid"
)

type CreateMentorAvailabilityRequest struct {
	DayOfWeek int    `json:"day_of_week" binding:"required,min=0,max=6"`
	StartTime string `json:"start_time" binding:"required"` // "10:00"
	EndTime   string `json:"end_time" binding:"required"`   // "18:00"
}

type MentorAvailabilityResponse struct {
	ID        uuid.UUID `json:"id"`
	DayOfWeek int       `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
}
