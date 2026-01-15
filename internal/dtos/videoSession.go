package dtos

import "github.com/google/uuid"

// WebSocket message types
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Start session request
type StartSessionRequest struct {
	BookingID string `json:"booking_id"` // UUID
	Role      string `json:"role"`       // "mentor" or "user"
	Timezone  string `json:"timezone"`   // e.g., "Asia/Kolkata"
}

// Session ready response
type SessionReadyResponse struct {
	BookingID       string `json:"booking_id"`
	OtherPartyRole  string `json:"other_party_role"` // "mentor" or "user"
	OtherPartyName  string `json:"other_party_name"`
	MaxDurationSecs int    `json:"max_duration_secs"` // 30 mins = 1800 secs
}

// Offer/Answer for WebRTC
type RTCOfferRequest struct {
	BookingID string `json:"booking_id"`
	SDP       string `json:"sdp"`
}

type RTCAnswerRequest struct {
	BookingID string `json:"booking_id"`
	SDP       string `json:"sdp"`
}

// ICE Candidate
type ICECandidateRequest struct {
	BookingID string `json:"booking_id"`
	Candidate string `json:"candidate"`
}

// Leave call
type LeaveCallRequest struct {
	BookingID string `json:"booking_id"`
}

// Session ended notification
type SessionEndedNotification struct {
	BookingID string `json:"booking_id"`
	Reason    string `json:"reason"`   // "time_limit_reached", "mentor_left", "user_left", "both_left"
	Duration  int    `json:"duration"` // seconds
}

// Video session response
type VideoSessionResponse struct {
	ID              uuid.UUID `json:"id"`
	BookingID       uuid.UUID `json:"booking_id"`
	MentorID        uuid.UUID `json:"mentor_id"`
	UserID          uuid.UUID `json:"user_id"`
	Status          string    `json:"status"`
	DurationSeconds int       `json:"duration_seconds"`
}

// Session info for join validation
type SessionInfoRequest struct {
	BookingID string `json:"booking_id"` // UUID
	Timezone  string `json:"timezone"`
}

type SessionInfoResponse struct {
	CanJoin          bool   `json:"can_join"`
	Message          string `json:"message"`
	StartTimeInTZ    string `json:"start_time_in_tz"` // formatted time
	EndTimeInTZ      string `json:"end_time_in_tz"`
	OtherPartyJoined bool   `json:"other_party_joined"`
	OtherPartyName   string `json:"other_party_name"`
}
