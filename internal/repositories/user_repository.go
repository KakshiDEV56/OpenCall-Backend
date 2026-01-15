package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/dtos"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

/*
ExistsByEmail checks whether a user already exists with given email
*/
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	const query = `
		SELECT 1
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

/*
ExistsByUsername checks whether a username already exists
*/
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	const query = `
		SELECT 1
		FROM users
		WHERE username = $1
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var exists int
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

/*
Create inserts a new user and returns the created row
*/
func (r *UserRepository) Create(user *models.User) (*models.User, error) {
	const query = `
		INSERT INTO users (
			id,
			first_name,
			last_name,
			username,
			email,
			password_hash,
			role,
			profile_picture,
			bio,
			is_active,
			created_at,
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
		RETURNING
			id,
			first_name,
			last_name,
			username,
			email,
			password_hash,
			role,
			profile_picture,
			bio,
			is_active,
			created_at,
			updated_at,
			deleted_at
	`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var created models.User
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.ProfilePicture,
		user.Bio,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&created.ID,
		&created.FirstName,
		&created.LastName,
		&created.Username,
		&created.Email,
		&created.PasswordHash,
		&created.Role,
		&created.ProfilePicture,
		&created.Bio,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *UserRepository) FindByEmailOrUsername(
	identifier string,
) (*models.User, error) {
	const query = `
		SELECT
			id,
			first_name,
			last_name,
			username,
			email,
			password_hash,
			role,
			profile_picture,
			bio,
			is_active,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE
			(email = $1 OR username = $1)
			AND deleted_at IS NULL
		LIMIT 1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User
	err := r.db.QueryRowContext(ctx, query, identifier).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.ProfilePicture,
		&user.Bio,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByID(userID uuid.UUID) (*models.User, error) {
	const query = `
		SELECT
			id,
			first_name,
			last_name,
			username,
			email,
			password_hash,
			role,
			profile_picture,
			bio,
			is_active,
			created_at,
			updated_at,
			deleted_at
		FROM users
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user models.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.ProfilePicture,
		&user.Bio,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindPublicProfileByUsername(
	username string,
) (*dtos.UserProfileResponse, error) {

	const query = `
	SELECT
		u.id,
		u.username,
		u.first_name,
		u.last_name,
		u.profile_picture,
		u.bio,

		(m.user_id IS NOT NULL) AS is_mentor,
		COALESCE(m.title, '')

	FROM users u
	LEFT JOIN mentor_profiles m ON m.user_id = u.id
	WHERE u.username = $1
	  AND u.deleted_at IS NULL
	LIMIT 1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var resp dtos.UserProfileResponse
	var mentorTitle string

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&resp.User.ID,
		&resp.User.Username,
		&resp.User.FirstName,
		&resp.User.LastName,
		&resp.User.ProfilePicture,
		&resp.User.Bio,
		&resp.IsMentor,
		&mentorTitle,
	)

	if err != nil {
		return nil, err
	}

	if resp.IsMentor {
		resp.Mentor = &dtos.MentorPreview{
			Title: mentorTitle,
		}
	}

	return &resp, nil
}
