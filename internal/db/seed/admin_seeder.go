package seed

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/goma-admin/internal/db/models"
	"github.com/jkaninda/goma-admin/internal/db/repository"

	"github.com/jkaninda/logger"
	"gorm.io/gorm"
)

type AdminConfig struct {
	Email    string
	Password string
	Name     string
	Username string
	Role     models.UserRole
}

func DefaultAdminConfig() *AdminConfig {
	return &AdminConfig{
		Email:    goutils.Env("GOMA_ADMIN_EMAIL", "admin@example.com"),
		Password: goutils.Env("GOMA_ADMIN_PASSWORD", "Admin@1234"),
		Name:     "Administrator",
		Username: "admin",
		Role:     models.RoleSuperAdmin,
	}
}

// CreateDefaultAdmin creates a default admin user if one doesn't exist
func CreateDefaultAdmin(db *gorm.DB) error {
	return CreateDefaultAdminWithConfig(db, DefaultAdminConfig())
}

func CreateDefaultAdminWithConfig(db *gorm.DB, config *AdminConfig) error {
	ctx := context.Background()
	repo := repository.NewUserRepository(db)

	empty, err := IsUsersTableEmpty(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to check users table: %w", err)
	}
	if !empty {
		logger.Info("Skipping admin seed: users table is not empty")
		return nil
	}

	// Check if admin already exists
	existingAdmin, err := repo.GetByEmail(ctx, config.Email)
	if err == nil && existingAdmin != nil {
		logger.Error("Admin user already exists", "Email", config.Email, "ID", existingAdmin.ID)
		return nil
	}

	// Check if username is taken
	if config.Username != "" {
		existsByUsername, _ := repo.ExistsByUsername(ctx, config.Username)
		if existsByUsername {
			return fmt.Errorf("username already exists: %s", config.Username)
		}
	}

	// Create admin user
	admin := &models.User{
		ID:            uuid.New(),
		Email:         config.Email,
		Name:          config.Name,
		Username:      config.Username,
		Role:          string(config.Role),
		EmailVerified: true,
		Active:        true,
		Metadata: models.JSONB{
			"created_by": "system",
			"is_seed":    true,
		},
	}

	// Set password
	if err := admin.SetPassword(config.Password); err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create user in database
	if err := repo.Create(ctx, admin); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		UserID:     admin.ID,
		Action:     "admin_created",
		Resource:   "user",
		ResourceID: admin.ID.String(),
		Status:     string(models.AuditStatusSuccess),
		Details: models.JSONB{
			"created_by": "system",
			"role":       config.Role,
			"seed":       true,
		},
	}

	if err := repo.CreateAuditLog(ctx, auditLog); err != nil {
		logger.Warn("Warning: Failed to create audit log for admin creation", "error", err)
	}

	logger.Info("Default admin user created successfully", "Email", config.Email, "Username", config.Username)

	if os.Getenv("ADMIN_PASSWORD") == "" {
		logger.Warn("Using default password. Please change it immediately!")
	}

	return nil
}
func IsUsersTableEmpty(ctx context.Context, db *gorm.DB) (bool, error) {
	var count int64
	if err := db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}
