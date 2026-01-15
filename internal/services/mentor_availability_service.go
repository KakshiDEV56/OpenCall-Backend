package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
)

type MentorAvailabilityService struct {
	availabilityRepo *repositories.MentorAvailabilityRepository
	mentorRepo       *repositories.MentorRepository
}

func NewMentorAvailabilityService(
	availabilityRepo *repositories.MentorAvailabilityRepository,
	mentorRepo *repositories.MentorRepository,
) *MentorAvailabilityService {
	return &MentorAvailabilityService{
		availabilityRepo: availabilityRepo,
		mentorRepo:       mentorRepo,
	}
}

func (s *MentorAvailabilityService) CreateRule(
	userID uuid.UUID,
	req *dtos.CreateMentorAvailabilityRequest,
) (*dtos.MentorAvailabilityResponse, error) {

	mentor, err := s.mentorRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	start, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, err
	}

	rule := &models.MentorAvailabilityRule{
		ID:        uuid.New(),
		MentorID:  mentor.ID,
		DayOfWeek: req.DayOfWeek,
		StartTime: start,
		EndTime:   end,
	}

	createdRule, err := s.availabilityRepo.Create(rule)
	if err != nil {
		return nil, err
	}

	return &dtos.MentorAvailabilityResponse{
		ID:        createdRule.ID,
		DayOfWeek: createdRule.DayOfWeek,
		StartTime: createdRule.StartTime.Format("15:04"),
		EndTime:   createdRule.EndTime.Format("15:04"),
	}, nil
}

func (s *MentorAvailabilityService) GetByUsername(
	username string,
) ([]dtos.MentorAvailabilityResponse, error) {

	rules, err := s.availabilityRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	resp := make([]dtos.MentorAvailabilityResponse, 0, len(rules))

	for _, r := range rules {
		resp = append(resp, dtos.MentorAvailabilityResponse{
			ID:        r.ID,
			DayOfWeek: r.DayOfWeek,
			StartTime: r.StartTime.Format("15:04"),
			EndTime:   r.EndTime.Format("15:04"),
		})
	}

	return resp, nil
}
