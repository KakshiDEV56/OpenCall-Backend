package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type VideoSessionRepository struct {
	db *sql.DB
}

func NewVideoSessionRepository(db *sql.DB) *VideoSessionRepository {
	return &VideoSessionRepository{db: db}
}

// Create a new video session
func (r *VideoSessionRepository) Create(ctx context.Context, session *models.VideoSession) error {
	const query = `
	INSERT INTO video_sessions (
		id,
		booking_id,
		mentor_id,
		user_id,
		status,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	RETURNING created_at, updated_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		session.ID,
		session.BookingID,
		session.MentorID,
		session.UserID,
		session.Status,
	).Scan(&session.CreatedAt, &session.UpdatedAt)
}

// Get session by booking ID
func (r *VideoSessionRepository) GetByBookingID(ctx context.Context, bookingID uuid.UUID) (*models.VideoSession, error) {
	const query = `
	SELECT
		id,
		booking_id,
		mentor_id,
		user_id,
		session_started_at,
		session_ended_at,
		mentor_joined_at,
		user_joined_at,
		mentor_left_at,
		user_left_at,
		duration_seconds,
		status,
		created_at,
		updated_at
	FROM video_sessions
	WHERE booking_id = $1
	LIMIT 1
	`

	var session models.VideoSession

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&session.ID,
		&session.BookingID,
		&session.MentorID,
		&session.UserID,
		&session.SessionStartedAt,
		&session.SessionEndedAt,
		&session.MentorJoinedAt,
		&session.UserJoinedAt,
		&session.MentorLeftAt,
		&session.UserLeftAt,
		&session.DurationSeconds,
		&session.Status,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("video session not found")
	}

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// Update session status
func (r *VideoSessionRepository) UpdateStatus(ctx context.Context, bookingID uuid.UUID, status models.VideoSessionStatus) error {
	const query = `
	UPDATE video_sessions
	SET status = $1, updated_at = NOW()
	WHERE booking_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, bookingID)
	return err
}

// Record when mentor joined
func (r *VideoSessionRepository) RecordMentorJoined(ctx context.Context, bookingID uuid.UUID) error {
	const query = `
	UPDATE video_sessions
	SET mentor_joined_at = NOW(), updated_at = NOW()
	WHERE booking_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, bookingID)
	return err
}

// Record when user joined
func (r *VideoSessionRepository) RecordUserJoined(ctx context.Context, bookingID uuid.UUID) error {
	const query = `
	UPDATE video_sessions
	SET user_joined_at = NOW(), updated_at = NOW()
	WHERE booking_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, bookingID)
	return err
}

// Record when mentor left
func (r *VideoSessionRepository) RecordMentorLeft(ctx context.Context, bookingID uuid.UUID) error {
	const query = `
	UPDATE video_sessions
	SET mentor_left_at = NOW(), updated_at = NOW()
	WHERE booking_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, bookingID)
	return err
}

// Record when user left
func (r *VideoSessionRepository) RecordUserLeft(ctx context.Context, bookingID uuid.UUID) error {
	const query = `
	UPDATE video_sessions
	SET user_left_at = NOW(), updated_at = NOW()
	WHERE booking_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, bookingID)
	return err
}

// End session and record duration
func (r *VideoSessionRepository) EndSession(ctx context.Context, bookingID uuid.UUID, durationSeconds int, reason string) error {
	const query = `
	UPDATE video_sessions
	SET 
		session_ended_at = NOW(),
		duration_seconds = $1,
		status = $2,
		updated_at = NOW()
	WHERE booking_id = $3
	`

	_, err := r.db.ExecContext(ctx, query, durationSeconds, models.VideoSessionStatusCompleted, bookingID)
	return err
}

// Mark session as active
func (r *VideoSessionRepository) MarkActive(ctx context.Context, bookingID uuid.UUID) error {
	const query = `
	UPDATE video_sessions
	SET 
		session_started_at = NOW(),
		status = $1,
		updated_at = NOW()
	WHERE booking_id = $2 AND status = $3
	`

	_, err := r.db.ExecContext(ctx, query, models.VideoSessionStatusActive, bookingID, models.VideoSessionStatusWaiting)
	return err
}
