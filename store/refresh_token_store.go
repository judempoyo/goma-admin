package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/models"
	"gorm.io/gorm"
)

type RefreshTokenStore interface {
	Create(ctx context.Context, token *models.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error
}

type GormRefreshTokenStore struct {
	db *gorm.DB
}

func NewRefreshTokenStore(db *gorm.DB) *GormRefreshTokenStore {
	return &GormRefreshTokenStore{db: db}
}

func (s *GormRefreshTokenStore) Create(ctx context.Context, token *models.RefreshToken) error {
	return s.db.WithContext(ctx).Create(token).Error
}

func (s *GormRefreshTokenStore) FindByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := s.db.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *GormRefreshTokenStore) Revoke(ctx context.Context, id uuid.UUID, revokedAt time.Time) error {
	return s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Updates(map[string]any{"revoked_at": revokedAt}).Error
}

func (s *GormRefreshTokenStore) RevokeAllForUser(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	return s.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{"revoked_at": revokedAt}).Error
}
