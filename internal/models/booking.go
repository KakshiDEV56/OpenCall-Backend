package models

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
)

type Booking struct {
	ID uuid.UUID `db:"id"`

	MentorID  uuid.UUID `db:"mentor_id"`
	UserID    uuid.UUID `db:"user_id"`
	ServiceID uuid.UUID `db:"service_id"`

	// Date & time are intentionally split
	BookingDate time.Time `db:"booking_date"` // DATE only
	StartTime   time.Time `db:"start_time"`   // TIME only
	EndTime     time.Time `db:"end_time"`     // TIME only

	Status BookingStatus `db:"status"`

	PriceCents int    `db:"price_cents"`
	Currency   string `db:"currency"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
