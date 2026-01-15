package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
)

type PaymentService struct {
	db             *sql.DB
	paymentRepo    *repositories.PaymentRepository
	bookingRepo    *repositories.BookingRepository
	razorpay       *RazorpayClient
	razorpaySecret string
}

func NewPaymentService(
	db *sql.DB,
	paymentRepo *repositories.PaymentRepository,
	bookingRepo *repositories.BookingRepository,
	razorpay *RazorpayClient,
	secret string,
) *PaymentService {
	return &PaymentService{
		db:             db,
		paymentRepo:    paymentRepo,
		bookingRepo:    bookingRepo,
		razorpay:       razorpay,
		razorpaySecret: secret,
	}
}

func (s *PaymentService) CreatePayment(
	ctx context.Context,
	bookingID uuid.UUID,
	userID uuid.UUID,
) (*models.Payment, error) {

	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	orderID, err := s.razorpay.CreateOrder(
		int64(booking.PriceCents),
		booking.ID.String(),
	)
	if err != nil {
		return nil, err
	}

	payment := &models.Payment{
		ID:             uuid.New(),
		BookingID:      booking.ID,
		UserID:         userID,
		Gateway:        "razorpay",
		GatewayOrderID: orderID,
		Amount:         int64(booking.PriceCents),
		Currency:       "INR",
		Status:         "created",
	}

	if err := s.paymentRepo.Create(ctx, tx, payment); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) VerifyPayment(
	ctx context.Context,
	paymentID uuid.UUID,
	razorpayPaymentID string,
	signature string,
) error {

	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		return err
	}

	data := payment.GatewayOrderID + "|" + razorpayPaymentID

	h := hmac.New(sha256.New, []byte(s.razorpaySecret))
	h.Write([]byte(data))
	expected := hex.EncodeToString(h.Sum(nil))

	if expected != signature {
		return errors.New("invalid payment signature")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.paymentRepo.MarkPaid(
		ctx,
		tx,
		paymentID,
		razorpayPaymentID,
		signature,
	); err != nil {
		return err
	}

	if err := s.bookingRepo.MarkConfirmed(ctx, tx, payment.BookingID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PaymentService) VerifyWebhookSignature(
	payload []byte,
	signature string,
) error {
	h := hmac.New(sha256.New, []byte(s.razorpaySecret))
	h.Write(payload)

	expected := hex.EncodeToString(h.Sum(nil))
	if expected != signature {
		return errors.New("invalid webhook signature")
	}
	return nil
}

func (s *PaymentService) HandlePaymentCaptured(
	event dtos.RazorpayWebhookEvent,
) error {

	orderID := event.Payload.Payment.Entity.OrderID
	paymentID := event.Payload.Payment.Entity.ID

	payment, err := s.paymentRepo.GetByGatewayOrderID(context.Background(), orderID)
	if err != nil {
		return err
	}

	if payment.Status == "paid" {
		return nil // idempotent
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.paymentRepo.MarkPaidByGateway(
		tx,
		payment.ID,
		paymentID,
	); err != nil {
		return err
	}

	if err := s.bookingRepo.MarkConfirmed(
		context.Background(),
		tx,
		payment.BookingID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PaymentService) HandlePaymentFailed(
	event dtos.RazorpayWebhookEvent,
) error {

	orderID := event.Payload.Payment.Entity.OrderID

	payment, err := s.paymentRepo.GetByGatewayOrderID(
		context.Background(),
		orderID,
	)
	if err != nil {
		return err
	}

	// Idempotency guard
	if payment.Status == "failed" || payment.Status == "paid" {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.paymentRepo.MarkFailedByGateway(
		tx,
		payment.ID,
	); err != nil {
		return err
	}

	if err := s.bookingRepo.MarkPaymentFailed(
		context.Background(),
		tx,
		payment.BookingID,
	); err != nil {
		return err
	}

	return tx.Commit()
}
