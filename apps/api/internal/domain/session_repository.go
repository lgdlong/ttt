package domain

import (
	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(session *Session) error
	GetByID(id uuid.UUID) (*Session, error)
	GetByRefreshToken(token string) (*Session, error)
	GetByUserID(userID uuid.UUID) ([]Session, error)
	GetActiveByUserID(userID uuid.UUID) ([]Session, error)
	Update(session *Session) error
	BlockSession(id uuid.UUID) error
	BlockAllUserSessions(userID uuid.UUID) error
	Delete(id uuid.UUID) error
	DeleteExpired() error
}
