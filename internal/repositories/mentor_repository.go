package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type MentorRepository struct {
	db *sql.DB
}

func NewMentorRepository(db *sql.DB) *MentorRepository {
	return &MentorRepository{db: db}
}

func (r *MentorRepository) CreateProfile(
	profile *models.MentorProfile,
) error {

	const query = `
	INSERT INTO mentor_profiles (
		id,
		user_id,
		title,
		bio,
		timezone,
		is_active,
		created_at,
		updated_at
	)
	VALUES ($1,$2,$3,$4,$5,true,NOW(),NOW())
	RETURNING created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.db.QueryRowContext(
		ctx,
		query,
		profile.ID,
		profile.UserID,
		profile.Title,
		profile.Bio,
		profile.Timezone,
	).Scan(&profile.CreatedAt, &profile.UpdatedAt)
}

func (r *MentorRepository) FindByUserID(
	userID uuid.UUID,
) (*models.MentorProfile, error) {

	const query = `
	SELECT
		id,
		user_id,
		title,
		bio,
		timezone,
		is_active,
		created_at,
		updated_at
	FROM mentor_profiles
	WHERE user_id = $1
	  AND is_active = true
	LIMIT 1
	`

	var mentor models.MentorProfile

	err := r.db.QueryRow(query, userID).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Title,
		&mentor.Bio,
		&mentor.Timezone,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &mentor, nil
}

func (r *MentorRepository) FindByUsername(
	username string,
) (*dtos.MentorProfileResponse, error) {

	const query = `
	SELECT
		u.id,
		u.username,
		u.first_name,
		u.last_name,
		u.profile_picture,

		m.id,
		m.title,
		m.bio,
		m.timezone,
		m.is_active
	FROM users u
	JOIN mentor_profiles m ON m.user_id = u.id
	WHERE u.username = $1
	  AND u.deleted_at IS NULL
	  AND m.is_active = true
	LIMIT 1
	`

	var resp dtos.MentorProfileResponse

	err := r.db.QueryRow(query, username).Scan(
		&resp.User.ID,
		&resp.User.Username,
		&resp.User.FirstName,
		&resp.User.LastName,
		&resp.User.ProfilePicture,

		&resp.Mentor.ID,
		&resp.Mentor.Title,
		&resp.Mentor.Bio,
		&resp.Mentor.Timezone,
		&resp.Mentor.IsActive,
	)

	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (r *MentorRepository) FindByUsernameRaw(
	username string,
) (*models.MentorProfile, error) {

	const query = `
	SELECT
		mp.id,
		mp.user_id,
		mp.title,
		mp.bio,
		mp.timezone,
		mp.is_active,
		mp.created_at,
		mp.updated_at
	FROM users u
	JOIN mentor_profiles mp ON mp.user_id = u.id
	WHERE u.username = $1
	  AND mp.is_active = true
	`

	var mentor models.MentorProfile

	err := r.db.QueryRow(
		query,
		username,
	).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Title,
		&mentor.Bio,
		&mentor.Timezone,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &mentor, nil
}

func (r *MentorRepository) FindByID(
	id uuid.UUID,
) (*models.MentorProfile, error) {

	const query = `
	SELECT
		id,
		user_id,
		title,
		bio,
		timezone,
		is_active,
		created_at,
		updated_at
	FROM mentor_profiles
	WHERE id = $1
	LIMIT 1
	`

	var mentor models.MentorProfile

	err := r.db.QueryRow(query, id).Scan(
		&mentor.ID,
		&mentor.UserID,
		&mentor.Title,
		&mentor.Bio,
		&mentor.Timezone,
		&mentor.IsActive,
		&mentor.CreatedAt,
		&mentor.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &mentor, nil
}
