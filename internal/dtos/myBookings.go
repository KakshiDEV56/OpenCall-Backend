package dtos

import "github.com/google/uuid"

type MyBookingResponse struct {
	ID        uuid.UUID `json:"id"`
	Mentor    string    `json:"mentor"` // username
	Service   string    `json:"service"`
	Date      string    `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Status    string    `json:"status"`
	Price     int       `json:"price_cents"`
	Currency  string    `json:"currency"`
}
