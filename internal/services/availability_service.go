package services

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
)

type AvailabilityService struct {
	mentorRepo       *repositories.MentorRepository
	serviceRepo      *repositories.MentorServiceRepository
	availabilityRepo *repositories.MentorAvailabilityRepository
	bookingRepo      *repositories.BookingRepository
}

func NewAvailabilityService(
	mentorRepo *repositories.MentorRepository,
	serviceRepo *repositories.MentorServiceRepository,
	availabilityRepo *repositories.MentorAvailabilityRepository,
	bookingRepo *repositories.BookingRepository,
) *AvailabilityService {
	return &AvailabilityService{
		mentorRepo:       mentorRepo,
		serviceRepo:      serviceRepo,
		availabilityRepo: availabilityRepo,
		bookingRepo:      bookingRepo,
	}
}

func (s *AvailabilityService) GetAvailableSlots(
	username string,
	serviceID uuid.UUID,
	dateStr string,
) (*dtos.AvailabilityResponse, error) {

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date")
	}

	mentor, err := s.mentorRepo.FindByUsernameRaw(username)
	if err != nil {
		return nil, err
	}

	service, err := s.serviceRepo.FindByID(serviceID)
	if err != nil || service.MentorID != mentor.ID {
		return nil, errors.New("invalid service")
	}

	dayOfWeek := int(date.Weekday())

	rules, err := s.availabilityRepo.FindByMentorAndDay(mentor.ID, dayOfWeek)
	if err != nil {
		return nil, err
	}

	bookings, err := s.bookingRepo.FindForMentorOnDate(
		mentor.ID,
		date,
	)
	if err != nil {
		return nil, err
	}

	duration := time.Duration(service.DurationMinutes) * time.Minute
	slots := []dtos.AvailableSlot{}

	for _, rule := range rules {
		start := time.Date(
			date.Year(), date.Month(), date.Day(),
			rule.StartTime.Hour(), rule.StartTime.Minute(), 0, 0, time.UTC,
		)

		end := time.Date(
			date.Year(), date.Month(), date.Day(),
			rule.EndTime.Hour(), rule.EndTime.Minute(), 0, 0, time.UTC,
		)

		for start.Add(duration).Equal(end) || start.Add(duration).Before(end) {
			slotEnd := start.Add(duration)

			if !overlaps(start, slotEnd, bookings) {
				slots = append(slots, dtos.AvailableSlot{
					Start: start.Format("15:04"),
					End:   slotEnd.Format("15:04"),
				})
			}

			start = slotEnd
		}
	}

	return &dtos.AvailabilityResponse{
		Date:  dateStr,
		Slots: slots,
	}, nil
}

func overlaps(start, end time.Time, bookings []*models.Booking) bool {
	for _, b := range bookings {

		bStart := time.Date(
			start.Year(), start.Month(), start.Day(),
			b.StartTime.Hour(), b.StartTime.Minute(), 0, 0, time.UTC,
		)

		bEnd := time.Date(
			start.Year(), start.Month(), start.Day(),
			b.EndTime.Hour(), b.EndTime.Minute(), 0, 0, time.UTC,
		)

		if start.Before(bEnd) && end.After(bStart) {
			return true
		}
	}
	return false
}
