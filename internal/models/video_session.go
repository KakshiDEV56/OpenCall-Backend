package models

import (
	"time"

	"github.com/google/uuid"
)

type VideoSessionStatus string

const (
	VideoSessionStatusWaiting   VideoSessionStatus = "waiting"
	VideoSessionStatusActive    VideoSessionStatus = "active"
	VideoSessionStatusCompleted VideoSessionStatus = "completed"
)

type VideoSession struct {
	ID        uuid.UUID `db:"id"`
	BookingID uuid.UUID `db:"booking_id"`
	MentorID  uuid.UUID `db:"mentor_id"`
	UserID    uuid.UUID `db:"user_id"`

	SessionStartedAt *time.Time `db:"session_started_at"`
	SessionEndedAt   *time.Time `db:"session_ended_at"`

	MentorJoinedAt *time.Time `db:"mentor_joined_at"`
	UserJoinedAt   *time.Time `db:"user_joined_at"`

	MentorLeftAt *time.Time `db:"mentor_left_at"`
	UserLeftAt   *time.Time `db:"user_left_at"`

	DurationSeconds int                `db:"duration_seconds"`
	Status          VideoSessionStatus `db:"status"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
