package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/models"
	"gorm.io/gorm"
)

type OAuthStore interface {
	FindByProviderID(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error)
	Create(ctx context.Context, account *models.OAuthAccount) error
	UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt *time.Time) error
}

type GormOAuthStore struct {
	db *gorm.DB
}

func NewOAuthStore(db *gorm.DB) *GormOAuthStore {
	return &GormOAuthStore{db: db}
}

func (s *GormOAuthStore) FindByProviderID(ctx context.Context, provider, providerUserID string) (*models.OAuthAccount, error) {
	var account models.OAuthAccount
	err := s.db.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *GormOAuthStore) Create(ctx context.Context, account *models.OAuthAccount) error {
	return s.db.WithContext(ctx).Create(account).Error
}

func (s *GormOAuthStore) UpdateTokens(ctx context.Context, id uuid.UUID, accessToken, refreshToken string, expiresAt *time.Time) error {
	updates := map[string]any{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    expiresAt,
	}
	return s.db.WithContext(ctx).
		Model(&models.OAuthAccount{}).
		Where("id = ?", id).
		Updates(updates).Error
}
