package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(
	ctx context.Context,
	tx *sql.Tx,
	p *models.Payment,
) error {

	query := `
		INSERT INTO payments (
			id,
			booking_id,
			user_id,
			gateway,
			gateway_order_id,
			amount,
			currency,
			status
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		p.ID,
		p.BookingID,
		p.UserID,
		p.Gateway,
		p.GatewayOrderID,
		p.Amount,
		p.Currency,
		p.Status,
	)

	return err
}

func (r *PaymentRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Payment, error) {

	query := `
		SELECT
			id,
			booking_id,
			user_id,
			gateway,
			gateway_order_id,
			gateway_payment_id,
			gateway_signature,
			amount,
			currency,
			status,
			created_at,
			updated_at
		FROM payments
		WHERE id = $1
	`

	var p models.Payment

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.BookingID,
		&p.UserID,
		&p.Gateway,
		&p.GatewayOrderID,
		&p.GatewayPaymentID,
		&p.GatewaySignature,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *PaymentRepository) MarkPaid(
	ctx context.Context,
	tx *sql.Tx,
	paymentID uuid.UUID,
	razorpayPaymentID string,
	signature string,
) error {

	query := `
		UPDATE payments
		SET
			status = 'paid',
			gateway_payment_id = $2,
			gateway_signature = $3,
			updated_at = now()
		WHERE id = $1
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		paymentID,
		razorpayPaymentID,
		signature,
	)

	return err
}

func (r *PaymentRepository) GetByGatewayOrderID(
	ctx context.Context,
	orderID string,
) (*models.Payment, error) {

	query := `
		SELECT
			id,
			booking_id,
			user_id,
			gateway,
			gateway_order_id,
			gateway_payment_id,
			gateway_signature,
			amount,
			currency,
			status,
			created_at,
			updated_at
		FROM payments
		WHERE gateway_order_id = $1
	`
	var p models.Payment

	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&p.ID,
		&p.BookingID,
		&p.UserID,
		&p.Gateway,
		&p.GatewayOrderID,
		&p.GatewayPaymentID,
		&p.GatewaySignature,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	return &p, err
}

func (r *PaymentRepository) MarkPaidByGateway(
	tx *sql.Tx,
	paymentID uuid.UUID,
	razorpayPaymentID string,
) error {

	query := `
		UPDATE payments
		SET
			status = 'paid',
			gateway_payment_id = $2,
			updated_at = now()
		WHERE id = $1
	`

	_, err := tx.Exec(
		query,
		paymentID,
		razorpayPaymentID,
	)

	return err
}

func (r *BookingRepository) MarkPaymentFailed(
	ctx context.Context,
	tx *sql.Tx,
	bookingID uuid.UUID,
) error {

	query := `
		UPDATE bookings
		SET status = 'payment_failed',
		    updated_at = now()
		WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, query, bookingID)
	return err
}

func (r *PaymentRepository) MarkFailedByGateway(
	tx *sql.Tx,
	paymentID uuid.UUID,
) error {

	query := `
		UPDATE payments
		SET
			status = 'failed',
			updated_at = now()
		WHERE id = $1
	`

	_, err := tx.Exec(query, paymentID)
	return err
}
