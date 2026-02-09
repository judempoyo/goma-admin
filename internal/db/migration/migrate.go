package migration

import (
	"fmt"

	"github.com/jkaninda/goma-admin/internal/db/models"
	"github.com/jkaninda/logger"
	"gorm.io/gorm"
)

// AutoMigrate runs all database migrations
func AutoMigrate(db *gorm.DB) error {
	logger.Info("Running database migrations...")
	err := db.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.AuditLog{},
		&models.Instance{},
		&models.Route{},
		&models.Backend{},
		&models.Maintenance{},
		&models.TLSCertificate{},
		&models.HealthCheck{},
		&models.Security{},
		&models.Middleware{},
		&models.RouteMiddleware{},
		&models.InstanceRoute{},
	)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Add custom indexes and constraints
	if err := addCustomIndexes(db); err != nil {
		return err
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

func addCustomIndexes(db *gorm.DB) error {
	if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_route_middleware_order 
        ON route_middlewares(route_id, execution_order)
    `).Error; err != nil {
		return err
	}

	if err := db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS idx_route_middleware_unique 
        ON route_middlewares(route_id, middleware_name)
    `).Error; err != nil {
		return err
	}

	if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_instance_routes_lookup
        ON instance_routes(instance_id, route_id)
    `).Error; err != nil {
		return err
	}

	if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_instances_health
        ON instances(enabled, status, last_seen)
    `).Error; err != nil {
		return err
	}

	return nil
}

func Rollback(db *gorm.DB) error {
	logger.Info("Rolling back database migrations...")

	err := db.Migrator().DropTable(
		&models.InstanceRoute{},
		&models.RouteMiddleware{},
		&models.Middleware{},
		&models.Security{},
		&models.HealthCheck{},
		&models.TLSCertificate{},
		&models.Maintenance{},
		&models.Backend{},
		&models.Route{},
		&models.Instance{},
	)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}
	logger.Info("Database rollback completed")
	return nil
}
