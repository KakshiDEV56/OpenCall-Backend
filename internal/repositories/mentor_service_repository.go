package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type MentorServiceRepository struct {
	db *sql.DB
}

func NewMentorServiceRepository(db *sql.DB) *MentorServiceRepository {
	return &MentorServiceRepository{db: db}
}

func (r *MentorServiceRepository) Create(
	service *models.MentorService,
) error {

	const query = `
	INSERT INTO mentor_services (
		id,
		mentor_id,
		title,
		description,
		duration_minutes,
		price_cents,
		currency,
		is_active,
		created_at,
		updated_at
	)
	VALUES ($1,$2,$3,$4,$5,$6,$7,true,NOW(),NOW())
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(
		ctx,
		query,
		service.ID,
		service.MentorID,
		service.Title,
		service.Description,
		service.DurationMinutes,
		service.PriceCents,
		service.Currency,
	)

	return err
}

func (r *MentorServiceRepository) FindByUsername(
	username string,
) ([]*models.MentorService, error) {

	const query = `
	SELECT
		ms.id,
		ms.mentor_id,
		ms.title,
		ms.description,
		ms.duration_minutes,
		ms.price_cents,
		ms.currency,
		ms.is_active,
		ms.created_at,
		ms.updated_at
	FROM users u
	JOIN mentor_profiles mp ON mp.user_id = u.id
	JOIN mentor_services ms ON ms.mentor_id = mp.id
	WHERE u.username = $1
	  AND u.deleted_at IS NULL
	  AND mp.is_active = true
	  AND ms.is_active = true
	ORDER BY ms.created_at ASC
	`

	rows, err := r.db.Query(query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*models.MentorService

	for rows.Next() {
		var s models.MentorService

		if err := rows.Scan(
			&s.ID,
			&s.MentorID,
			&s.Title,
			&s.Description,
			&s.DurationMinutes,
			&s.PriceCents,
			&s.Currency,
			&s.IsActive,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}

		services = append(services, &s)
	}

	return services, nil
}

func (r *MentorServiceRepository) FindByID(
	serviceID uuid.UUID,
) (*models.MentorService, error) {

	const query = `
	SELECT
		id,
		mentor_id,
		title,
		description,
		duration_minutes,
		price_cents,
		currency,
		is_active,
		created_at,
		updated_at
	FROM mentor_services
	WHERE id = $1
	  AND is_active = true
	`

	var service models.MentorService

	err := r.db.QueryRow(
		query,
		serviceID,
	).Scan(
		&service.ID,
		&service.MentorID,
		&service.Title,
		&service.Description,
		&service.DurationMinutes,
		&service.PriceCents,
		&service.Currency,
		&service.IsActive,
		&service.CreatedAt,
		&service.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &service, nil
}
