package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
)

type MentorAvailabilityHandler struct {
	mentorAvailabilityService *services.MentorAvailabilityService
	availabilityService       *services.AvailabilityService
}

func NewMentorAvailabilityHandler(
	mentorAvailabilityService *services.MentorAvailabilityService,
	availabilityService *services.AvailabilityService,
) *MentorAvailabilityHandler {
	return &MentorAvailabilityHandler{
		mentorAvailabilityService: mentorAvailabilityService,
		availabilityService:       availabilityService,
	}
}

func (h *MentorAvailabilityHandler) Create(c *gin.Context) {
	var req dtos.CreateMentorAvailabilityRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("user_id")
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	resp, err := h.mentorAvailabilityService.CreateRule(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot create availability"})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *MentorAvailabilityHandler) GetByUsername(c *gin.Context) {
	username := c.Param("username")

	// Check if query parameters are present for available slots
	date := c.Query("date")
	serviceIDStr := c.Query("service_id")

	if date != "" && serviceIDStr != "" {
		// Return available slots for a specific date and service
		serviceID, err := uuid.Parse(serviceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service_id"})
			return
		}

		resp, err := h.availabilityService.GetAvailableSlots(username, serviceID, date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
		return
	}

	// Return all availability rules for the mentor
	resp, err := h.mentorAvailabilityService.GetByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "availability not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *MentorAvailabilityHandler) Get(c *gin.Context) {
	username := c.Param("username")
	date := c.Query("date")
	serviceIDStr := c.Query("service_id")

	serviceID, err := uuid.Parse(serviceIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid service_id"})
		return
	}

	resp, err := h.availabilityService.GetAvailableSlots(username, serviceID, date)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, resp)
}
