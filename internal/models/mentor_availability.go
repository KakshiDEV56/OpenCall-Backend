package models

import (
	"time"

	"github.com/google/uuid"
)

type MentorAvailabilityRule struct {
	ID        uuid.UUID
	MentorID  uuid.UUID
	DayOfWeek int
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
