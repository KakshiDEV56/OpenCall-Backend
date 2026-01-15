package dtos

import "github.com/google/uuid"

// Step 1: create Razorpay order
type CreatePaymentRequest struct {
	BookingID uuid.UUID `json:"booking_id" binding:"required"`
}

type CreatePaymentResponse struct {
	PaymentID       uuid.UUID `json:"payment_id"`
	RazorpayOrderID string    `json:"razorpay_order_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
}

// Step 2: verify payment
type VerifyPaymentRequest struct {
	PaymentID         uuid.UUID `json:"payment_id" binding:"required"`
	RazorpayPaymentID string    `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string    `json:"razorpay_signature" binding:"required"`
}
