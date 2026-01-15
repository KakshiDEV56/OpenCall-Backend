package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/services"
)

type MentorHandler struct {
	mentorProfileService *services.MentorProfileService
}

func NewMentorHandler(
	mentorProfileService *services.MentorProfileService,
) *MentorHandler {
	return &MentorHandler{
		mentorProfileService: mentorProfileService,
	}
}

func (h *MentorHandler) CreateProfile(c *gin.Context) {
	var req dtos.CreateMentorProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	profile, err := h.mentorProfileService.CreateProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "mentor profile already exists or other error",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.CreateMentorProfileResponse{
		ID:        profile.ID,
		UserID:    profile.UserID,
		Title:     profile.Title,
		Bio:       profile.Bio,
		Timezone:  profile.Timezone,
		IsActive:  profile.IsActive,
		CreatedAt: profile.CreatedAt,
	})
}

func (h *MentorHandler) GetProfile(c *gin.Context) {
	username := c.Param("username")

	resp, err := h.mentorProfileService.GetMentorProfile(username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "mentor not found",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}
