package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
)

type BookingHandler struct {
	bookingService *services.BookingService
	mentorRepo     *repositories.MentorRepository
}

func NewBookingHandler(
	bookingService *services.BookingService,
	mentorRepo *repositories.MentorRepository,
) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		mentorRepo:     mentorRepo,
	}
}

func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req dtos.CreateBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user id",
		})
		return
	}

	resp, err := h.bookingService.CreateBooking(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *BookingHandler) GetMyBookings(c *gin.Context) {

	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	resp, err := h.bookingService.GetMyBookings(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *BookingHandler) GetMentorBookedSessions(c *gin.Context) {

	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	// Find mentor profile for this user
	mentor, err := h.mentorRepo.FindByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mentor profile not found"})
		return
	}

	// Get booked sessions for this mentor
	sessions, err := h.bookingService.GetMentorBookedSessions(mentor.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch booked sessions"})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
