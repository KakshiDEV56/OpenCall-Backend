package websocket

import (
	"encoding/json"
	"sync"
)

// MessageBuffer buffers WebRTC signaling messages until both clients are ready
type MessageBuffer struct {
	mu       sync.Mutex
	messages []WebSocketMessage
	maxSize  int
}

// WebSocketMessage is the standard message format for all WebSocket communication
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SessionReadyPayload is sent when both participants have joined
type SessionReadyPayload struct {
	OtherPartyName string `json:"other_party_name"`
	Role           string `json:"role"` // "mentor" or "user"
}

// RTCOfferPayload contains SDP offer
type RTCOfferPayload struct {
	SDP string `json:"sdp"`
}

// RTCAnswerPayload contains SDP answer
type RTCAnswerPayload struct {
	SDP string `json:"sdp"`
}

// ICECandidatePayload contains full ICE candidate data
type ICECandidatePayload struct {
	Candidate     string  `json:"candidate"`     // The candidate string
	SDPMid        *string `json:"sdpMid"`        // Media stream id, can be null
	SDPMLineIndex *int    `json:"sdpMLineIndex"` // Media line index, can be null
}

// LeaveCallPayload is sent when user leaves the call
type LeaveCallPayload struct{}

// NewMessageBuffer creates a new message buffer
func NewMessageBuffer(maxSize int) *MessageBuffer {
	return &MessageBuffer{
		messages: make([]WebSocketMessage, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add adds a message to the buffer
// Returns error if buffer is full
func (mb *MessageBuffer) Add(msg WebSocketMessage) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if len(mb.messages) >= mb.maxSize {
		return ErrMessageBufferFull
	}

	mb.messages = append(mb.messages, msg)
	return nil
}

// Flush returns all buffered messages and clears the buffer
func (mb *MessageBuffer) Flush() []WebSocketMessage {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	messages := mb.messages
	mb.messages = make([]WebSocketMessage, 0, mb.maxSize)
	return messages
}

// IsEmpty returns true if buffer has no messages
func (mb *MessageBuffer) IsEmpty() bool {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return len(mb.messages) == 0
}

// Size returns current buffer size
func (mb *MessageBuffer) Size() int {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	return len(mb.messages)
}

// ConnectionState tracks the state of a WebSocket connection
type ConnectionState struct {
	mu                  sync.RWMutex
	PeerConnectionReady bool
	SessionReadySent    bool
	MessageBufferActive bool
	messageBuffer       *MessageBuffer
}

// NewConnectionState creates a new connection state
func NewConnectionState() *ConnectionState {
	return &ConnectionState{
		PeerConnectionReady: false,
		SessionReadySent:    false,
		MessageBufferActive: true,
		messageBuffer:       NewMessageBuffer(100),
	}
}

// IsPeerConnectionReady returns true if peer connection is initialized
func (cs *ConnectionState) IsPeerConnectionReady() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.PeerConnectionReady
}

// SetPeerConnectionReady marks peer connection as ready
func (cs *ConnectionState) SetPeerConnectionReady(ready bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.PeerConnectionReady = ready
	if ready {
		cs.MessageBufferActive = false
	}
}

// HasSessionReadySent returns true if session_ready was already sent
func (cs *ConnectionState) HasSessionReadySent() bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.SessionReadySent
}

// SetSessionReadySent marks session_ready as sent
func (cs *ConnectionState) SetSessionReadySent(sent bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.SessionReadySent = sent
}

// BufferMessage adds a message to the buffer if buffering is active
// Returns true if message was buffered, false if should be processed immediately
func (cs *ConnectionState) BufferMessage(msg WebSocketMessage) bool {
	cs.mu.Lock()
	isBuffering := cs.MessageBufferActive
	cs.mu.Unlock()

	if !isBuffering {
		return false
	}

	err := cs.messageBuffer.Add(msg)
	return err == nil
}

// FlushBuffer returns all buffered messages
func (cs *ConnectionState) FlushBuffer() []WebSocketMessage {
	return cs.messageBuffer.Flush()
}
