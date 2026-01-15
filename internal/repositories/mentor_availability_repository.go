package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type MentorAvailabilityRepository struct {
	db *sql.DB
}

func NewMentorAvailabilityRepository(db *sql.DB) *MentorAvailabilityRepository {
	return &MentorAvailabilityRepository{db: db}
}

func (r *MentorAvailabilityRepository) Create(
	rule *models.MentorAvailabilityRule,
) (*models.MentorAvailabilityRule, error) {

	const query = `
	INSERT INTO mentor_availability_rules (
		id,
		mentor_id,
		day_of_week,
		start_time,
		end_time,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(
		ctx,
		query,
		rule.ID,
		rule.MentorID,
		rule.DayOfWeek,
		rule.StartTime,
		rule.EndTime,
	)
	if err != nil {
		return nil, err
	}

	return rule, nil
}

func (r *MentorAvailabilityRepository) FindByUsername(
	username string,
) ([]*models.MentorAvailabilityRule, error) {

	const query = `
	SELECT
		ar.id,
		ar.day_of_week,
		ar.start_time,
		ar.end_time
	FROM users u
	JOIN mentor_profiles mp ON mp.user_id = u.id
	JOIN mentor_availability_rules ar ON ar.mentor_id = mp.id
	WHERE u.username = $1
	  AND mp.is_active = true
	ORDER BY ar.day_of_week, ar.start_time
	`

	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.MentorAvailabilityRule

	for rows.Next() {
		var rule models.MentorAvailabilityRule

		if err := rows.Scan(
			&rule.ID,
			&rule.DayOfWeek,
			&rule.StartTime,
			&rule.EndTime,
		); err != nil {
			return nil, err
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}

func (r *MentorAvailabilityRepository) FindByMentorAndDay(
	mentorID uuid.UUID,
	dayOfWeek int,
) ([]*models.MentorAvailabilityRule, error) {

	const query = `
	SELECT
		id,
		day_of_week,
		start_time,
		end_time
	FROM mentor_availability_rules
	WHERE mentor_id = $1
	  AND day_of_week = $2
	ORDER BY start_time
	`

	rows, err := r.db.Query(query, mentorID, dayOfWeek)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*models.MentorAvailabilityRule

	for rows.Next() {
		var rule models.MentorAvailabilityRule

		if err := rows.Scan(
			&rule.ID,
			&rule.DayOfWeek,
			&rule.StartTime,
			&rule.EndTime,
		); err != nil {
			return nil, err
		}

		rules = append(rules, &rule)
	}

	return rules, nil
}
