package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	ID uuid.UUID `db:"id"`

	BookingID uuid.UUID `db:"booking_id"`
	UserID    uuid.UUID `db:"user_id"`

	Gateway string `db:"gateway"`

	GatewayOrderID   string  `db:"gateway_order_id"`
	GatewayPaymentID *string `db:"gateway_payment_id"`
	GatewaySignature *string `db:"gateway_signature"`

	Amount   int64  `db:"amount"`
	Currency string `db:"currency"`

	Status string `db:"status"` // created | paid | failed

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
