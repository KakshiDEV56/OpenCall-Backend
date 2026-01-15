package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/preetsinghmakkar/OpenCall/internal/models"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) FindByTokenHash(
	tokenHash string,
) (*models.RefreshToken, error) {

	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var rt models.RefreshToken

	err := r.db.QueryRow(query, tokenHash).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.TokenHash,
		&rt.ExpiresAt,
		&rt.RevokedAt,
	)

	return &rt, err
}

func (r *RefreshTokenRepository) Create(
	userID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) error {

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(query, userID, tokenHash, expiresAt)
	return err
}

func (r *RefreshTokenRepository) Revoke(id uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked_at IS NULL
	`
	_, err := r.db.Exec(query, userID)
	return err
}
