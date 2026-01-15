package models

import (
	"time"

	"github.com/google/uuid"
)

type MentorProfile struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	Bio       string    `json:"bio" db:"bio"`
	Timezone  string    `json:"timezone" db:"timezone"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
