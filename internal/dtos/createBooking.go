package dtos

import "github.com/google/uuid"

// --------------------
// CREATE BOOKING
// --------------------

type CreateBookingRequest struct {
	MentorID    uuid.UUID `json:"mentor_id" binding:"required"`
	ServiceID   uuid.UUID `json:"service_id" binding:"required"`
	BookingDate string    `json:"booking_date" binding:"required"` // YYYY-MM-DD
	StartTime   string    `json:"start_time" binding:"required"`   // HH:MM
}

// --------------------
// RESPONSE
// --------------------

type BookingResponse struct {
	ID        uuid.UUID `json:"id"`
	Status    string    `json:"status"`
	Date      string    `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Price     int       `json:"price_cents"`
	Currency  string    `json:"currency"`
}
