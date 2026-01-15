package websocket

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	ID              uuid.UUID
	BookingID       uuid.UUID
	Role            string // "mentor" or "user" - derived from booking ownership
	UserID          uuid.UUID
	MentorID        uuid.UUID // For reference
	Username        string
	Conn            *websocket.Conn
	Hub             *Hub
	Send            chan interface{}
	Done            chan struct{}
	ConnectionState *ConnectionState // Tracks signaling state
}

// Hub manages all active WebSocket connections for a booking
type Hub struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*Session // key: booking_id
}

// Session represents an active video session with both parties
type Session struct {
	BookingID   uuid.UUID
	Mentor      *Client
	User        *Client
	StartTime   time.Time
	MaxDuration int // seconds
	Done        chan struct{}
	mu          sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		sessions: make(map[uuid.UUID]*Session),
	}
}

// NewSession creates a new session
func NewSession(bookingID uuid.UUID, maxDuration int) *Session {
	return &Session{
		BookingID:   bookingID,
		StartTime:   time.Now(),
		MaxDuration: maxDuration,
		Done:        make(chan struct{}),
	}
}

// AddClient adds a client to a session
// If a client already exists for this (user, booking), close the old one first
func (h *Hub) AddClient(bookingID uuid.UUID, client *Client) *Session {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[bookingID]
	if !exists {
		session = NewSession(bookingID, 30*60) // 30 minutes
		h.sessions[bookingID] = session
	}

	// Close old connection if it exists (prevent duplicate connections)
	if client.Role == "mentor" {
		if session.Mentor != nil && session.Mentor.ID != client.ID {
			fmt.Println("[Hub] Closing duplicate mentor connection for booking", bookingID)
			session.Mentor.Close()
		}
		session.Mentor = client
	} else {
		if session.User != nil && session.User.ID != client.ID {
			fmt.Println("[Hub] Closing duplicate user connection for booking", bookingID)
			session.User.Close()
		}
		session.User = client
	}

	return session
}

// GetSession gets a session by booking ID
func (h *Hub) GetSession(bookingID uuid.UUID) *Session {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.sessions[bookingID]
}

// RemoveClient removes a client from a session
func (h *Hub) RemoveClient(bookingID uuid.UUID, role string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	session, exists := h.sessions[bookingID]
	if !exists {
		return
	}

	if role == "mentor" {
		session.Mentor = nil
	} else {
		session.User = nil
	}

	// If both are gone, remove session
	if session.Mentor == nil && session.User == nil {
		delete(h.sessions, bookingID)
	}
}

// BothJoined checks if both parties have joined
func (s *Session) BothJoined() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Mentor != nil && s.User != nil
}

// GetOtherClient gets the other party
func (s *Session) GetOtherClient(role string) *Client {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if role == "mentor" {
		return s.User
	}
	return s.Mentor
}

// Broadcast sends a message to both parties
func (s *Session) Broadcast(message interface{}) {
	s.mu.RLock()
	mentor := s.Mentor
	user := s.User
	s.mu.RUnlock()

	if mentor != nil {
		select {
		case mentor.Send <- message:
		default:
		}
	}

	if user != nil {
		select {
		case user.Send <- message:
		default:
		}
	}
}

// SendToRole sends a message to a specific role
func (s *Session) SendToRole(role string, message interface{}) {
	s.mu.RLock()
	var client *Client
	if role == "mentor" {
		client = s.Mentor
	} else {
		client = s.User
	}
	s.mu.RUnlock()

	if client != nil {
		select {
		case client.Send <- message:
		default:
		}
	}
}

// Close closes the session
func (s *Session) Close() {
	close(s.Done)

	s.mu.RLock()
	mentor := s.Mentor
	user := s.User
	s.mu.RUnlock()

	if mentor != nil {
		mentor.Close()
	}
	if user != nil {
		user.Close()
	}
}

// Close closes the client connection
func (c *Client) Close() {
	close(c.Send)
	c.Conn.Close()
	close(c.Done)
}

// IsConnected checks if client is still connected
func (c *Client) IsConnected() bool {
	select {
	case <-c.Done:
		return false
	default:
		return true
	}
}
