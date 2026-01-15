package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
	"github.com/preetsinghmakkar/OpenCall/internal/utils"
)

// WebSocketAuthContext holds authenticated WebSocket connection data
type WebSocketAuthContext struct {
	UserID    uuid.UUID
	Username  string
	BookingID uuid.UUID
	Role      string // "mentor" or "user"
	MentorID  uuid.UUID
	Booking   *models.Booking
}

// WebSocketAuthMiddleware authenticates WebSocket connections
// Validates JWT, booking ownership, and derives role from database
// Must be used BEFORE WebSocket upgrade
func WebSocketAuthMiddleware(
	jwtSecret string,
	bookingRepo *repositories.BookingRepository,
	// Used to load usernames and mentor profiles
	userRepo *repositories.UserRepository,
	mentorRepo *repositories.MentorRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract and validate JWT token
		token := c.Query("token")
		if token == "" {
			fmt.Println("[WebSocket Auth] ERROR: Token missing from query parameters")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		claims, err := utils.ParseAccessToken(token, jwtSecret)
		if err != nil {
			fmt.Println("[WebSocket Auth] ERROR: JWT validation failed:", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		userID := claims.UserID
		fmt.Println("[WebSocket Auth] JWT validated, UserID:", userID)

		// Extract and validate booking_id
		bookingIDStr := c.Query("booking_id")
		if bookingIDStr == "" {
			fmt.Println("[WebSocket Auth] ERROR: booking_id missing")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "booking_id required",
			})
			return
		}

		bookingID, err := uuid.Parse(bookingIDStr)
		if err != nil {
			fmt.Println("[WebSocket Auth] ERROR: Invalid booking_id format:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "invalid booking_id format",
			})
			return
		}

		// Load booking from database
		ctx := c.Request.Context()
		booking, err := bookingRepo.GetByID(ctx, bookingID)
		if err != nil {
			fmt.Println("[WebSocket Auth] ERROR: Failed to load booking:", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "booking not found",
			})
			return
		}

		// Derive role from booking ownership (NEVER trust client).
		// For users, bookings.user_id stores the user ID.
		// For mentors, bookings.mentor_id stores the mentor_profile ID,
		// so we need to resolve the mentor profile for this user.
		var role string

		// Try mentor ownership first
		if mentorRepo != nil {
			if mentorProfile, err := mentorRepo.FindByUserID(userID); err == nil && mentorProfile != nil {
				if booking.MentorID == mentorProfile.ID {
					role = "mentor"
				}
			}
		}

		// If not mentor for this booking, check if this is the user
		if role == "" && booking.UserID == userID {
			role = "user"
		}

		if role == "" {
			fmt.Printf("[WebSocket Auth] ERROR: User %s not authorized for booking %s\n", userID, bookingID)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "not authorized for this booking",
			})
			return
		}

		fmt.Println("[WebSocket Auth] Role derived from booking:", role)

		// Load username from database (NEVER trust client)
		user, err := userRepo.FindByID(userID)
		if err != nil {
			fmt.Println("[WebSocket Auth] ERROR: Failed to load user:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "user not found",
			})
			return
		}

		if user == nil || user.Username == "" {
			fmt.Println("[WebSocket Auth] ERROR: User has no username")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "invalid user profile",
			})
			return
		}

		fmt.Println("[WebSocket Auth] Username loaded from database:", user.Username)

		// Create authentication context
		authCtx := &WebSocketAuthContext{
			UserID:    userID,
			Username:  user.Username,
			BookingID: bookingID,
			Role:      role,
			MentorID:  booking.MentorID,
			Booking:   booking,
		}

		// Store in context for handler to access
		ctx = context.WithValue(c.Request.Context(), "ws_auth", authCtx)
		c.Request = c.Request.WithContext(ctx)

		fmt.Println("[WebSocket Auth] ✓ Authentication successful")
		fmt.Printf("[WebSocket Auth] UserID=%s, BookingID=%s, Role=%s, Username=%s\n",
			userID, bookingID, role, user.Username)

		c.Next()
	}
}

// GetWebSocketAuth retrieves authentication context from request
func GetWebSocketAuth(c *gin.Context) (*WebSocketAuthContext, error) {
	val := c.Request.Context().Value("ws_auth")
	if val == nil {
		return nil, errors.New("websocket authentication context not found")
	}

	auth, ok := val.(*WebSocketAuthContext)
	if !ok {
		return nil, errors.New("invalid websocket authentication context type")
	}

	return auth, nil
}
