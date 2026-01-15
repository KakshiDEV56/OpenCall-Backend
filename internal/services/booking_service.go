package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
	"github.com/preetsinghmakkar/OpenCall/internal/repositories"
)

type BookingService struct {
	bookingRepo      *repositories.BookingRepository
	mentorRepo       *repositories.MentorRepository
	serviceRepo      *repositories.MentorServiceRepository
	availabilityRepo *repositories.MentorAvailabilityRepository
}

func NewBookingService(
	bookingRepo *repositories.BookingRepository,
	mentorRepo *repositories.MentorRepository,
	serviceRepo *repositories.MentorServiceRepository,
	availabilityRepo *repositories.MentorAvailabilityRepository,
) *BookingService {
	return &BookingService{
		bookingRepo:      bookingRepo,
		mentorRepo:       mentorRepo,
		serviceRepo:      serviceRepo,
		availabilityRepo: availabilityRepo,
	}
}

func (s *BookingService) CreateBooking(
	userID uuid.UUID,
	req *dtos.CreateBookingRequest,
) (*dtos.BookingResponse, error) {

	// 1️⃣ Parse date
	bookingDate, err := time.Parse("2006-01-02", req.BookingDate)
	if err != nil {
		return nil, errors.New("invalid date format")
	}

	// 2️⃣ Parse start time
	startTimeParsed, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, errors.New("invalid start time")
	}

	// 3️⃣ Fetch service
	service, err := s.serviceRepo.FindByID(req.ServiceID)
	if err != nil || !service.IsActive {
		return nil, errors.New("invalid service")
	}

	// 4️⃣ Fetch mentor
	mentor, err := s.mentorRepo.FindByID(service.MentorID)
	if err != nil || !mentor.IsActive {
		return nil, errors.New("mentor not available")
	}

	// 5️⃣ Build start/end datetime (UTC)
	start := time.Date(
		bookingDate.Year(), bookingDate.Month(), bookingDate.Day(),
		startTimeParsed.Hour(), startTimeParsed.Minute(),
		0, 0, time.UTC,
	)

	duration := time.Duration(service.DurationMinutes) * time.Minute
	end := start.Add(duration)

	// 6️⃣ Validate availability rules
	dayOfWeek := int(bookingDate.Weekday())

	rules, err := s.availabilityRepo.FindByMentorAndDay(mentor.ID, dayOfWeek)
	if err != nil || len(rules) == 0 {
		return nil, errors.New("mentor not available on this day")
	}

	valid := false
	for _, r := range rules {
		ruleStart := time.Date(
			start.Year(), start.Month(), start.Day(),
			r.StartTime.Hour(), r.StartTime.Minute(),
			0, 0, time.UTC,
		)

		ruleEnd := time.Date(
			start.Year(), start.Month(), start.Day(),
			r.EndTime.Hour(), r.EndTime.Minute(),
			0, 0, time.UTC,
		)

		if !start.Before(ruleStart) && !end.After(ruleEnd) {
			valid = true
			break
		}
	}

	if !valid {
		return nil, errors.New("selected slot outside availability")
	}

	// 7️⃣ TRANSACTION START
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	bookingID := uuid.New()

	err = s.bookingRepo.WithTx(ctx, func(tx *sql.Tx) error {

		conflict, err := s.bookingRepo.HasConflictTx(
			ctx,
			tx,
			mentor.ID,
			bookingDate,
			start,
			end,
		)
		if err != nil {
			return err
		}

		if conflict {
			return errors.New("slot already booked")
		}

		booking := &models.Booking{
			ID:          bookingID,
			MentorID:    mentor.ID,
			UserID:      userID,
			ServiceID:   service.ID,
			BookingDate: bookingDate,
			StartTime:   start,
			EndTime:     end,
			Status:      models.BookingStatusPending,
			PriceCents:  service.PriceCents,
			Currency:    service.Currency,
		}

		return s.bookingRepo.CreateTx(ctx, tx, booking)
	})

	if err != nil {
		return nil, err
	}

	// Response
	return &dtos.BookingResponse{
		ID:        bookingID,
		Status:    string(models.BookingStatusPending),
		Date:      req.BookingDate,
		StartTime: start.Format("15:04"),
		EndTime:   end.Format("15:04"),
		Price:     service.PriceCents,
		Currency:  service.Currency,
	}, nil
}

func (s *BookingService) GetMyBookings(
	userID uuid.UUID,
) ([]*dtos.MyBookingResponse, error) {

	bookings, err := s.bookingRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	if bookings == nil {
		return []*dtos.MyBookingResponse{}, nil
	}

	return bookings, nil
}

func (s *BookingService) GetMentorBookedSessions(
	mentorID uuid.UUID,
) ([]*dtos.MentorBookedSessionResponse, error) {

	sessions, err := s.bookingRepo.GetByMentorIDConfirmed(mentorID)
	if err != nil {
		return nil, err
	}

	if sessions == nil {
		return []*dtos.MentorBookedSessionResponse{}, nil
	}

	return sessions, nil
}
