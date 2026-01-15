package services

import (
	"strings"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
)

type MentorProfileService struct {
	mentorRepo *repositories.MentorRepository
}

func NewMentorProfileService(
	mentorRepo *repositories.MentorRepository,
) *MentorProfileService {
	return &MentorProfileService{
		mentorRepo: mentorRepo,
	}
}

func (s *MentorProfileService) CreateProfile(
	userID uuid.UUID,
	req *dtos.CreateMentorProfileRequest,
) (*models.MentorProfile, error) {

	profile := &models.MentorProfile{
		ID:       uuid.New(),
		UserID:   userID,
		Title:    strings.TrimSpace(req.Title),
		Bio:      strings.TrimSpace(req.Bio),
		Timezone: req.Timezone,
		IsActive: true,
	}

	if err := s.mentorRepo.CreateProfile(profile); err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *MentorProfileService) GetMentorProfile(
	username string,
) (*dtos.MentorProfileResponse, error) {
	return s.mentorRepo.FindByUsername(username)
}
