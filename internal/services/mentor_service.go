package services

import (
	"strings"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
)

type MentorOfferingService struct {
	serviceRepo *repositories.MentorServiceRepository
	mentorRepo  *repositories.MentorRepository
}

func NewMentorOfferingService(
	serviceRepo *repositories.MentorServiceRepository,
	mentorRepo *repositories.MentorRepository,
) *MentorOfferingService {
	return &MentorOfferingService{
		serviceRepo: serviceRepo,
		mentorRepo:  mentorRepo,
	}
}

func (s *MentorOfferingService) CreateService(
	userID uuid.UUID,
	req *dtos.CreateMentorServiceRequest,
) (*models.MentorService, error) {

	mentor, err := s.mentorRepo.FindByUserID(userID)
	if err != nil {
		return nil, err // mentor profile not found
	}

	service := &models.MentorService{
		ID:              uuid.New(),
		MentorID:        mentor.ID,
		Title:           strings.TrimSpace(req.Title),
		Description:     strings.TrimSpace(req.Description),
		DurationMinutes: req.DurationMinutes,
		PriceCents:      req.PriceCents,
		Currency:        strings.ToUpper(req.Currency),
		IsActive:        true,
	}
	if err := s.serviceRepo.Create(service); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *MentorOfferingService) GetServicesByUsername(
	username string,
) ([]dtos.MentorServiceResponse, error) {

	services, err := s.serviceRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	resp := make([]dtos.MentorServiceResponse, 0, len(services))

	for _, svc := range services {
		resp = append(resp, dtos.MentorServiceResponse{
			ID:              svc.ID,
			Title:           svc.Title,
			Description:     svc.Description,
			DurationMinutes: svc.DurationMinutes,
			PriceCents:      svc.PriceCents,
			Currency:        svc.Currency,
			IsActive:        svc.IsActive,
		})
	}

	return resp, nil
}
