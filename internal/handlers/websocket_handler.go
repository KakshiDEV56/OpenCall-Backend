package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/preetsinghmakkar/OpenCall/internal/middlewares"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
	ws "github.com/preetsinghmakkar/OpenCall/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now, restrict in production
	},
}

type WebSocketHandler struct {
	videoSessionService *services.VideoSessionService
	hub                 *ws.Hub
	jwtSecret           string
}

func NewWebSocketHandler(
	videoSessionService *services.VideoSessionService,
	hub *ws.Hub,
	jwtSecret string,
) *WebSocketHandler {
	return &WebSocketHandler{
		videoSessionService: videoSessionService,
		hub:                 hub,
		jwtSecret:           jwtSecret,
	}
}

// HandleWebSocket is the WebSocket endpoint handler
// MUST be protected by WebSocketAuthMiddleware
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Retrieve authenticated context set by middleware
	auth, err := middlewares.GetWebSocketAuth(c)
	if err != nil {
		fmt.Println("[WebSocket Handler] ERROR: Missing authentication context:", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	fmt.Println("[WebSocket Handler] Authentication context retrieved")
	fmt.Printf("[WebSocket Handler] UserID=%s, BookingID=%s, Role=%s, Username=%s\n",
		auth.UserID, auth.BookingID, auth.Role, auth.Username)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("[WebSocket Handler] ERROR: WebSocket upgrade failed:", err)
		return
	}

	fmt.Println("[WebSocket Handler] ✓ WebSocket upgrade successful")

	// Create client with authenticated data
	client := &ws.Client{
		ID:              uuid.New(),
		BookingID:       auth.BookingID,
		Role:            auth.Role,
		UserID:          auth.UserID,
		MentorID:        auth.MentorID,
		Username:        auth.Username,
		Conn:            conn,
		Hub:             h.hub,
		Send:            make(chan interface{}, 256),
		Done:            make(chan struct{}),
		ConnectionState: ws.NewConnectionState(),
	}

	// Add client to hub and create session if needed
	session := h.hub.AddClient(auth.BookingID, client)

	fmt.Println("[WebSocket Handler] Client added to session")

	// Record client joined in database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	h.videoSessionService.HandleClientJoined(ctx, client, auth.Username)
	cancel()

	// Send session_ready when both parties are present
	h.sendSessionReady(session, client, auth)

	// Start reading and writing goroutines
	go h.readPump(client, session)
	go h.writePump(client)
}

// sendSessionReady sends session_ready message to both clients when both have joined
func (h *WebSocketHandler) sendSessionReady(session *ws.Session, newClient *ws.Client, auth *middlewares.WebSocketAuthContext) {
	if !session.BothJoined() {
		fmt.Println("[WebSocket Handler] Waiting for other party to join...")
		return
	}

	// Get the other client
	otherClient := session.GetOtherClient(auth.Role)
	if otherClient == nil {
		fmt.Println("[WebSocket Handler] ERROR: Other client is nil despite BothJoined being true")
		return
	}

	fmt.Println("[WebSocket Handler] Both parties joined, sending session_ready")

	// Get other party's username
	otherUsername := otherClient.Username

	// Build and send message to new client
	if !newClient.ConnectionState.HasSessionReadySent() {
		sessionReadyMsg := ws.WebSocketMessage{
			Type: "session_ready",
			Payload: func() json.RawMessage {
				payload := ws.SessionReadyPayload{
					OtherPartyName: otherUsername,
					Role:           auth.Role,
				}
				data, _ := json.Marshal(payload)
				return data
			}(),
		}

		select {
		case newClient.Send <- map[string]interface{}{
			"type":    sessionReadyMsg.Type,
			"payload": json.RawMessage(sessionReadyMsg.Payload),
		}:
			newClient.ConnectionState.SetSessionReadySent(true)
			fmt.Println("[WebSocket Handler] Sent session_ready to", auth.Role)
		default:
			fmt.Println("[WebSocket Handler] WARNING: Failed to send session_ready to", auth.Role)
		}
	}

	// Build and send message to other client (if not already sent)
	if !otherClient.ConnectionState.HasSessionReadySent() {
		otherSessionReadyMsg := ws.WebSocketMessage{
			Type: "session_ready",
			Payload: func() json.RawMessage {
				payload := ws.SessionReadyPayload{
					OtherPartyName: auth.Username,
					Role:           otherClient.Role,
				}
				data, _ := json.Marshal(payload)
				return data
			}(),
		}

		select {
		case otherClient.Send <- map[string]interface{}{
			"type":    otherSessionReadyMsg.Type,
			"payload": json.RawMessage(otherSessionReadyMsg.Payload),
		}:
			otherClient.ConnectionState.SetSessionReadySent(true)
			fmt.Println("[WebSocket Handler] Sent session_ready to", otherClient.Role)
		default:
			fmt.Println("[WebSocket Handler] WARNING: Failed to send session_ready to", otherClient.Role)
		}
	}
}

// readPump reads messages from the WebSocket and forwards them appropriately
func (h *WebSocketHandler) readPump(client *ws.Client, session *ws.Session) {
	defer func() {
		fmt.Printf("[WebSocket ReadPump] Cleaning up client %s (role=%s)\n", client.UserID, client.Role)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		h.videoSessionService.HandleClientLeft(ctx, client.BookingID, client.Role)
		cancel()

		h.hub.RemoveClient(client.BookingID, client.Role)
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var rawMsg json.RawMessage
		err := client.Conn.ReadJSON(&rawMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("[WebSocket ReadPump] Unexpected close error: %v\n", err)
			}
			return
		}

		// Parse message envelope (type + payload)
		var msg ws.WebSocketMessage
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			fmt.Printf("[WebSocket ReadPump] ERROR: Failed to parse message: %v\n", err)
			continue
		}

		fmt.Printf("[WebSocket ReadPump] Received message type=%s from %s\n", msg.Type, client.Role)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		switch msg.Type {
		case "offer":
			h.handleOfferMessage(client, session, msg.Payload, ctx)

		case "answer":
			h.handleAnswerMessage(client, session, msg.Payload, ctx)

		case "ice_candidate":
			h.handleICECandidateMessage(client, session, msg.Payload, ctx)

		case "leave_call":
			fmt.Printf("[WebSocket ReadPump] Leave call from %s\n", client.Role)
			cancel()
			return

		case "ping":
			fmt.Printf("[WebSocket ReadPump] Ping from %s\n", client.Role)
			response := map[string]interface{}{
				"type":    "pong",
				"payload": map[string]interface{}{},
			}
			select {
			case client.Send <- response:
			default:
			}

		default:
			fmt.Printf("[WebSocket ReadPump] WARNING: Unknown message type: %s\n", msg.Type)
		}

		cancel()
	}
}

// handleOfferMessage handles incoming SDP offer
func (h *WebSocketHandler) handleOfferMessage(
	client *ws.Client,
	session *ws.Session,
	payload json.RawMessage,
	ctx context.Context,
) {
	var offerPayload ws.RTCOfferPayload
	if err := json.Unmarshal(payload, &offerPayload); err != nil {
		fmt.Printf("[WebSocket] ERROR: Failed to parse offer: %v\n", err)
		return
	}

	if offerPayload.SDP == "" {
		fmt.Println("[WebSocket] ERROR: Empty SDP in offer")
		return
	}

	fmt.Printf("[WebSocket] Forwarding offer from %s to %s\n", client.Role,
		map[string]string{"mentor": "user", "user": "mentor"}[client.Role])

	// Forward to other party
	response := map[string]interface{}{
		"type":    "offer",
		"payload": offerPayload,
	}

	otherRole := "user"
	if client.Role == "user" {
		otherRole = "mentor"
	}
	session.SendToRole(otherRole, response)
}

// handleAnswerMessage handles incoming SDP answer
func (h *WebSocketHandler) handleAnswerMessage(
	client *ws.Client,
	session *ws.Session,
	payload json.RawMessage,
	ctx context.Context,
) {
	var answerPayload ws.RTCAnswerPayload
	if err := json.Unmarshal(payload, &answerPayload); err != nil {
		fmt.Printf("[WebSocket] ERROR: Failed to parse answer: %v\n", err)
		return
	}

	if answerPayload.SDP == "" {
		fmt.Println("[WebSocket] ERROR: Empty SDP in answer")
		return
	}

	fmt.Printf("[WebSocket] Forwarding answer from %s to %s\n", client.Role,
		map[string]string{"mentor": "user", "user": "mentor"}[client.Role])

	// Forward to other party
	response := map[string]interface{}{
		"type":    "answer",
		"payload": answerPayload,
	}

	otherRole := "user"
	if client.Role == "user" {
		otherRole = "mentor"
	}
	session.SendToRole(otherRole, response)
}

// handleICECandidateMessage handles incoming ICE candidate
func (h *WebSocketHandler) handleICECandidateMessage(
	client *ws.Client,
	session *ws.Session,
	payload json.RawMessage,
	ctx context.Context,
) {
	var icePayload ws.ICECandidatePayload
	if err := json.Unmarshal(payload, &icePayload); err != nil {
		fmt.Printf("[WebSocket] ERROR: Failed to parse ICE candidate: %v\n", err)
		return
	}

	if icePayload.Candidate == "" {
		fmt.Println("[WebSocket] WARNING: Empty ICE candidate")
		return
	}

	fmt.Printf("[WebSocket] Forwarding ICE candidate from %s to %s\n", client.Role,
		map[string]string{"mentor": "user", "user": "mentor"}[client.Role])

	// Forward to other party
	response := map[string]interface{}{
		"type":    "ice_candidate",
		"payload": icePayload,
	}

	otherRole := "user"
	if client.Role == "user" {
		otherRole = "mentor"
	}
	session.SendToRole(otherRole, response)
}

// writePump writes messages to the WebSocket
func (h *WebSocketHandler) writePump(client *ws.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		fmt.Printf("[WebSocket WritePump] Stopping for client %s\n", client.UserID)
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.Conn.WriteJSON(message); err != nil {
				fmt.Printf("[WebSocket WritePump] ERROR: Failed to write message: %v\n", err)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("[WebSocket WritePump] ERROR: Failed to send ping: %v\n", err)
				return
			}

		case <-client.Done:
			fmt.Printf("[WebSocket WritePump] Client done for %s\n", client.UserID)
			return
		}
	}
}

// GetSessionInfo returns information about a video session
func (h *WebSocketHandler) GetSessionInfo(c *gin.Context) {
	bookingIDStr := c.Query("booking_id")
	timezone := c.Query("timezone")

	if bookingIDStr == "" || timezone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "booking_id and timezone required"})
		return
	}

	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid booking_id"})
		return
	}

	// Get user ID from auth middleware context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		fmt.Println("[GetSessionInfo] ERROR: user_id not in context - auth middleware may not have run")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	userIDValue, ok := userIDStr.(string)
	if !ok {
		fmt.Println("[GetSessionInfo] ERROR: user_id is not a string:", userIDStr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id format"})
		return
	}

	userID, err := uuid.Parse(userIDValue)
	if err != nil {
		fmt.Println("[GetSessionInfo] ERROR: failed to parse user_id:", userIDValue, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user_id"})
		return
	}

	// Determine role - try to find if user is mentor or regular user
	role := "user"
	// You might want to check mentor_profiles table here
	// For now, we'll accept role from query param or default to user

	if queryRole := c.Query("role"); queryRole != "" {
		role = queryRole
	}

	fmt.Printf("[GetSessionInfo] User: %s, Booking: %s, Role: %s, Timezone: %s\n", userID, bookingID, role, timezone)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessionInfo, err := h.videoSessionService.ValidateAndStartSession(ctx, bookingID, userID, role, timezone)
	if err != nil {
		fmt.Println("[GetSessionInfo] Session validation error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessionInfo)
}
