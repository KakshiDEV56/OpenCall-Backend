package dtos

import "github.com/google/uuid"

type MentorBookedSessionResponse struct {
	ID           uuid.UUID `json:"id"`
	UserUsername string    `json:"user_username"`
	ServiceTitle string    `json:"service_title"`
	BookingDate  string    `json:"booking_date"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	PriceCents   int       `json:"price_cents"`
	Currency     string    `json:"currency"`
}
