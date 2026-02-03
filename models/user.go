package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email         string         `gorm:"uniqueIndex;not null"`
	PasswordHash  string         `gorm:"column:password_hash"`
	Name          string         `gorm:""`
	EmailVerified bool           `gorm:"default:false"`
	Roles         []Role         `gorm:"many2many:user_roles;"`
	OAuthAccounts []OAuthAccount `gorm:"constraint:OnDelete:CASCADE"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
