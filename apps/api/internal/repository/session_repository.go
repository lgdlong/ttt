package repository

import (
	"api/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) domain.SessionRepository {
	return &sessionRepository{db: db}
}

// Create creates a new session
func (r *sessionRepository) Create(session *domain.Session) error {
	return r.db.Create(session).Error
}

// GetByID retrieves a session by ID
func (r *sessionRepository) GetByID(id uuid.UUID) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.Preload("User").First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByRefreshToken retrieves a session by refresh token
func (r *sessionRepository) GetByRefreshToken(token string) (*domain.Session, error) {
	var session domain.Session
	if err := r.db.Preload("User").
		Where("refresh_token = ?", token).
		First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserID retrieves all sessions for a user
func (r *sessionRepository) GetByUserID(userID uuid.UUID) ([]domain.Session, error) {
	var sessions []domain.Session
	if err := r.db.Where("user_id = ?", userID).Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetActiveByUserID retrieves all active (non-blocked, non-expired) sessions for a user
func (r *sessionRepository) GetActiveByUserID(userID uuid.UUID) ([]domain.Session, error) {
	var sessions []domain.Session
	if err := r.db.Where("user_id = ? AND is_blocked = ? AND expires_at > ?",
		userID, false, time.Now()).
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// Update updates a session
func (r *sessionRepository) Update(session *domain.Session) error {
	return r.db.Save(session).Error
}

// BlockSession blocks a specific session
func (r *sessionRepository) BlockSession(id uuid.UUID) error {
	return r.db.Model(&domain.Session{}).
		Where("id = ?", id).
		Update("is_blocked", true).Error
}

// BlockAllUserSessions blocks all sessions for a user
func (r *sessionRepository) BlockAllUserSessions(userID uuid.UUID) error {
	return r.db.Model(&domain.Session{}).
		Where("user_id = ?", userID).
		Update("is_blocked", true).Error
}

// Delete removes a session
func (r *sessionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Session{}, "id = ?", id).Error
}

// DeleteExpired removes all expired sessions
func (r *sessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&domain.Session{}).Error
}
