package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
)

type MentorServiceHandler struct {
	service *services.MentorOfferingService
}

func NewMentorServiceHandler(
	service *services.MentorOfferingService,
) *MentorServiceHandler {
	return &MentorServiceHandler{service: service}
}

func (h *MentorServiceHandler) Create(c *gin.Context) {
	var req dtos.CreateMentorServiceRequest

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

	service, err := h.service.CreateService(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "mentor profile not found",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.MentorServiceResponse{
		ID:              service.ID,
		Title:           service.Title,
		Description:     service.Description,
		DurationMinutes: service.DurationMinutes,
		PriceCents:      service.PriceCents,
		Currency:        service.Currency,
		IsActive:        service.IsActive,
	})
}

func (h *MentorServiceHandler) GetByUsername(c *gin.Context) {
	username := c.Param("username")

	services, err := h.service.GetServicesByUsername(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "mentor services not found",
		})
		return
	}

	c.JSON(http.StatusOK, services)
}
