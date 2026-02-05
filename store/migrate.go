package store

import (
	"errors"

	"github.com/jkaninda/goma-admin/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("database is nil")
	}
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.UserRole{},
		&models.RefreshToken{},
		&models.OAuthAccount{},
	)
}
