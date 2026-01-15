package utils

import (
	"errors"
	"time"
)

// ValidateSessionTime checks if current time is within (or just around)
// the booking window for the given timezone. A small grace window is
// allowed so users can join a few minutes before the scheduled start
// and remain connected briefly after the scheduled end.
func ValidateSessionTime(bookingStartTime, bookingEndTime time.Time, timezone string) (bool, string, error) {
	// Load the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return false, "", errors.New("invalid timezone")
	}

	// Get current time in user's timezone
	now := time.Now().In(loc)
	startInTZ := bookingStartTime.In(loc)
	endInTZ := bookingEndTime.In(loc)

	// Allow a small grace window before and after the scheduled time
	graceBefore := 5 * time.Minute
	graceAfter := 5 * time.Minute

	joinWindowStart := startInTZ.Add(-graceBefore)
	joinWindowEnd := endInTZ.Add(graceAfter)

	// Check if current time is before join window
	if now.Before(joinWindowStart) {
		return false, "Session has not started yet. Available at " + startInTZ.Format("15:04 MST"), nil
	}

	// Check if current time is after allowed end window
	if now.After(joinWindowEnd) {
		return false, "Session time has ended", nil
	}

	// Within valid (or grace) window
	return true, "", nil
}

// FormatTimeInTimezone formats a time in the given timezone
func FormatTimeInTimezone(t time.Time, timezone string) (string, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}

	return t.In(loc).Format("15:04 MST"), nil
}

// GetMaxSessionDuration returns the duration in seconds (30 minutes)
func GetMaxSessionDuration() int {
	return 30 * 60 // 1800 seconds
}

// CalculateDuration calculates duration between two times in seconds
func CalculateDuration(start, end time.Time) int {
	return int(end.Sub(start).Seconds())
}
