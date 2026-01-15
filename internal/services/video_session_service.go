package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
	"github.com/preetsinghmakkar/OpenCall/internal/utils"
	"github.com/preetsinghmakkar/OpenCall/internal/websocket"
)

type VideoSessionService struct {
	videoSessionRepo *repositories.VideoSessionRepository
	bookingRepo      *repositories.BookingRepository
	mentorRepo       *repositories.MentorRepository
	userRepo         *repositories.UserRepository
	hub              *websocket.Hub
}

func NewVideoSessionService(
	videoSessionRepo *repositories.VideoSessionRepository,
	bookingRepo *repositories.BookingRepository,
	mentorRepo *repositories.MentorRepository,
	userRepo *repositories.UserRepository,
	hub *websocket.Hub,
) *VideoSessionService {
	return &VideoSessionService{
		videoSessionRepo: videoSessionRepo,
		bookingRepo:      bookingRepo,
		mentorRepo:       mentorRepo,
		userRepo:         userRepo,
		hub:              hub,
	}
}

// ValidateAndStartSession validates booking time and creates/returns video session
func (s *VideoSessionService) ValidateAndStartSession(
	ctx context.Context,
	bookingID uuid.UUID,
	userID uuid.UUID,
	role string,
	timezone string,
) (*dtos.SessionInfoResponse, error) {
	// Get booking
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	// Verify user owns this booking
	if role == "user" && booking.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	// For mentor, verify they own the booking.
	// bookings.mentor_id stores the mentor_profile ID, while the JWT
	// contains the underlying user ID. Resolve the mentor profile for
	// this user and compare its ID to the booking's mentor ID.
	if role == "mentor" {
		mentorProfile, err := s.mentorRepo.FindByUserID(userID)
		if err != nil || mentorProfile == nil || booking.MentorID != mentorProfile.ID {
			return nil, errors.New("unauthorized")
		}
	}

	// Check booking status is confirmed
	if booking.Status != models.BookingStatusConfirmed {
		return nil, errors.New("booking must be confirmed")
	}

	// Combine booking date with start/end times to get full datetimes
	// BookingDate stores the calendar date, StartTime/EndTime store the time of day.
	date := booking.BookingDate
	start := booking.StartTime
	end := booking.EndTime

	// Interpret stored times as being in the user's timezone (e.g. IST),
	// not in UTC. If we used date.Location() here (often UTC), then later
	// converting to the user timezone would incorrectly shift the clock
	// (e.g. 15:47 local becoming 21:17 IST).
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, errors.New("invalid timezone")
	}
	bookingStart := time.Date(
		date.Year(), date.Month(), date.Day(),
		start.Hour(), start.Minute(), start.Second(), start.Nanosecond(),
		loc,
	)
	bookingEnd := time.Date(
		date.Year(), date.Month(), date.Day(),
		end.Hour(), end.Minute(), end.Second(), end.Nanosecond(),
		loc,
	)

	// Validate timezone and time window using full datetimes
	canJoin, msg, err := utils.ValidateSessionTime(bookingStart, bookingEnd, timezone)
	if err != nil {
		return nil, errors.New("invalid timezone")
	}

	if !canJoin {
		return &dtos.SessionInfoResponse{
			CanJoin: false,
			Message: msg,
		}, nil
	}

	// Format times in user's timezone
	startStr, _ := utils.FormatTimeInTimezone(bookingStart, timezone)
	endStr, _ := utils.FormatTimeInTimezone(bookingEnd, timezone)

	// Check if video session already exists
	existingSession, _ := s.videoSessionRepo.GetByBookingID(ctx, bookingID)

	otherPartyJoined := false
	otherPartyName := ""

	if existingSession != nil {
		// Session exists, check if other party joined
		session := s.hub.GetSession(bookingID)
		if session != nil {
			otherParty := session.GetOtherClient(role)
			if otherParty != nil {
				otherPartyJoined = true
				otherPartyName = otherParty.Username
			}
		}
	} else {
		// Create new video session
		videoSession := &models.VideoSession{
			ID:        uuid.New(),
			BookingID: bookingID,
			MentorID:  booking.MentorID,
			UserID:    booking.UserID,
			Status:    models.VideoSessionStatusWaiting,
		}

		if err := s.videoSessionRepo.Create(ctx, videoSession); err != nil {
			return nil, errors.New("failed to create video session")
		}
	}

	return &dtos.SessionInfoResponse{
		CanJoin:          true,
		Message:          "",
		StartTimeInTZ:    startStr,
		EndTimeInTZ:      endStr,
		OtherPartyJoined: otherPartyJoined,
		OtherPartyName:   otherPartyName,
	}, nil
}

// HandleClientJoined handles when a client joins the WebSocket
// username is provided from authenticated context (from database, not client)
func (s *VideoSessionService) HandleClientJoined(
	ctx context.Context,
	client *websocket.Client,
	username string,
) error {
	// Record in database
	if client.Role == "mentor" {
		s.videoSessionRepo.RecordMentorJoined(ctx, client.BookingID)
	} else {
		s.videoSessionRepo.RecordUserJoined(ctx, client.BookingID)
	}

	return nil
}

// HandleClientLeft handles when a client disconnects
func (s *VideoSessionService) HandleClientLeft(ctx context.Context, bookingID uuid.UUID, role string) error {
	// Record in database
	if role == "mentor" {
		s.videoSessionRepo.RecordMentorLeft(ctx, bookingID)
	} else {
		s.videoSessionRepo.RecordUserLeft(ctx, bookingID)
	}

	// Get session
	session := s.hub.GetSession(bookingID)
	if session != nil {
		// Notify other party if present
		otherParty := session.GetOtherClient(role)
		if otherParty != nil {
			response := map[string]interface{}{
				"type": "other_party_left",
				"payload": map[string]string{
					"booking_id": bookingID.String(),
				},
			}

			select {
			case otherParty.Send <- response:
			default:
			}

			// Close other party's connection
			otherParty.Close()
		}

		// End the session
		s.EndSession(ctx, bookingID, "party_left")
	}

	// Remove from hub
	s.hub.RemoveClient(bookingID, role)

	return nil
}

// ForwardMessage forwards a WebRTC message to the other party
func (s *VideoSessionService) ForwardMessage(ctx context.Context, bookingID uuid.UUID, senderRole string, messageType string, payload interface{}) error {
	session := s.hub.GetSession(bookingID)
	if session == nil {
		return errors.New("session not found")
	}

	response := map[string]interface{}{
		"type":    messageType,
		"payload": payload,
	}

	otherRole := "user"
	if senderRole == "user" {
		otherRole = "mentor"
	}

	session.SendToRole(otherRole, response)
	return nil
}

// EndSession marks session as completed and updates booking
func (s *VideoSessionService) EndSession(ctx context.Context, bookingID uuid.UUID, reason string) error {
	// Get session to calculate duration
	session, err := s.videoSessionRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return err
	}

	// Calculate duration
	duration := 0
	if session.SessionStartedAt != nil && session.SessionEndedAt == nil {
		duration = int(time.Since(*session.SessionStartedAt).Seconds())
	}

	// End video session
	err = s.videoSessionRepo.EndSession(ctx, bookingID, duration, reason)
	if err != nil {
		return err
	}

	// Do not change booking status here.
	// Booking remains "confirmed" so it stays visible in
	// mentor/user booking lists; time-window validation already
	// prevents joining outside the allowed period.
	return nil
}

// GetSessionInfo returns info about a video session
func (s *VideoSessionService) GetSessionInfo(ctx context.Context, bookingID uuid.UUID) (*dtos.VideoSessionResponse, error) {
	session, err := s.videoSessionRepo.GetByBookingID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	return &dtos.VideoSessionResponse{
		ID:              session.ID,
		BookingID:       session.BookingID,
		MentorID:        session.MentorID,
		UserID:          session.UserID,
		Status:          string(session.Status),
		DurationSeconds: session.DurationSeconds,
	}, nil
}

// MarshalWebSocketMessage converts a message to JSON
func MarshalWebSocketMessage(messageType string, payload interface{}) ([]byte, error) {
	msg := map[string]interface{}{
		"type":    messageType,
		"payload": payload,
	}
	return json.Marshal(msg)
}
