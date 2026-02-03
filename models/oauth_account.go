package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OAuthAccount struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Provider       string     `gorm:"size:32;index:idx_oauth_provider_user,unique"`
	ProviderUserID string     `gorm:"size:128;index:idx_oauth_provider_user,unique"`
	UserID         uuid.UUID  `gorm:"type:uuid;index"`
	AccessToken    string     `gorm:""`
	RefreshToken   string     `gorm:""`
	ExpiresAt      *time.Time `gorm:""`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (o *OAuthAccount) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}
