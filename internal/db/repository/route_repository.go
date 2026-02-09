package repository

import (
	"context"
	"fmt"

	"github.com/jkaninda/goma-admin/internal/db/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RouteRepository struct {
	db *gorm.DB
}

func NewRouteRepository(db *gorm.DB) *RouteRepository {
	return &RouteRepository{db: db}
}

// Create creates a new route with all its associations
func (r *RouteRepository) Create(ctx context.Context, route *models.Route) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(route).Error; err != nil {
			return fmt.Errorf("failed to create route: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a route by ID with all associations
func (r *RouteRepository) GetByID(ctx context.Context, id uint) (*models.Route, error) {
	var route models.Route

	err := r.db.WithContext(ctx).
		Preload("Backends").
		Preload("Maintenance").
		Preload("TLSCertificates").
		Preload("HealthCheck").
		Preload("Security").
		Preload("RouteMiddlewares", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_middlewares.execution_order ASC")
		}).
		First(&route, id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("route not found: %d", id)
		}
		return nil, err
	}

	return &route, nil
}

// GetByName retrieves a route by name with all associations
func (r *RouteRepository) GetByName(ctx context.Context, name string) (*models.Route, error) {
	var route models.Route

	err := r.db.WithContext(ctx).
		Preload("Backends").
		Preload("Maintenance").
		Preload("TLSCertificates").
		Preload("HealthCheck").
		Preload("Security").
		Preload("RouteMiddlewares", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_middlewares.execution_order ASC")
		}).
		Where("name = ?", name).
		First(&route).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("route not found: %s", name)
		}
		return nil, err
	}

	return &route, nil
}

// List retrieves all routes with their associations
func (r *RouteRepository) List(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route

	err := r.db.WithContext(ctx).
		Preload("Backends").
		Preload("Maintenance").
		Preload("TLSCertificates").
		Preload("HealthCheck").
		Preload("Security").
		Preload("RouteMiddlewares", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_middlewares.execution_order ASC")
		}).
		Order("priority DESC, name ASC").
		Find(&routes).Error

	if err != nil {
		return nil, err
	}

	return routes, nil
}

// ListEnabled retrieves only enabled routes
func (r *RouteRepository) ListEnabled(ctx context.Context) ([]models.Route, error) {
	var routes []models.Route

	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Preload("Backends").
		Preload("Maintenance").
		Preload("TLSCertificates").
		Preload("HealthCheck").
		Preload("Security").
		Preload("RouteMiddlewares", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_middlewares.execution_order ASC")
		}).
		Order("priority DESC, name ASC").
		Find(&routes).Error

	if err != nil {
		return nil, err
	}

	return routes, nil
}

// Update updates a route and its associations
func (r *RouteRepository) Update(ctx context.Context, route *models.Route) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the route basic fields
		if err := tx.Model(route).Updates(map[string]interface{}{
			"path":            route.Path,
			"rewrite":         route.Rewrite,
			"priority":        route.Priority,
			"enabled":         route.Enabled,
			"methods":         route.Methods,
			"hosts":           route.Hosts,
			"target":          route.Target,
			"disable_metrics": route.DisableMetrics,
		}).Error; err != nil {
			return fmt.Errorf("failed to update route: %w", err)
		}

		// Replace backends
		if err := tx.Where("route_id = ?", route.ID).Delete(&models.Backend{}).Error; err != nil {
			return err
		}
		if len(route.Backends) > 0 {
			for i := range route.Backends {
				route.Backends[i].RouteID = route.ID
			}
			if err := tx.Create(&route.Backends).Error; err != nil {
				return err
			}
		}

		// Update or create maintenance
		if route.Maintenance != nil {
			route.Maintenance.RouteID = route.ID
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "route_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"enabled", "status_code", "message"}),
			}).Create(route.Maintenance).Error; err != nil {
				return err
			}
		} else {
			tx.Where("route_id = ?", route.ID).Delete(&models.Maintenance{})
		}

		// Replace TLS certificates - handle both TLS wrapper and TLSCertificates
		if err := tx.Where("route_id = ?", route.ID).Delete(&models.TLSCertificate{}).Error; err != nil {
			return err
		}

		var certs []models.TLSCertificate
		if route.TLS != nil && len(route.TLS.Certificates) > 0 {
			certs = route.TLS.Certificates
		} else if len(route.TLSCertificates) > 0 {
			certs = route.TLSCertificates
		}

		if len(certs) > 0 {
			for i := range certs {
				certs[i].RouteID = route.ID
			}
			if err := tx.Create(&certs).Error; err != nil {
				return err
			}
		}

		// Update or create health check
		if route.HealthCheck != nil {
			route.HealthCheck.RouteID = route.ID
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "route_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"path", "interval", "timeout", "healthy_statuses"}),
			}).Create(route.HealthCheck).Error; err != nil {
				return err
			}
		} else {
			tx.Where("route_id = ?", route.ID).Delete(&models.HealthCheck{})
		}

		// Update or create security
		if route.Security != nil {
			route.Security.RouteID = route.ID
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "route_id"}},
				DoUpdates: clause.AssignmentColumns([]string{
					"forward_host_headers", "enable_exploit_protection",
					"tls_insecure_skip_verify", "tls_root_cas", "tls_client_cert", "tls_client_key",
				}),
			}).Create(route.Security).Error; err != nil {
				return err
			}
		} else {
			tx.Where("route_id = ?", route.ID).Delete(&models.Security{})
		}

		// Replace route middlewares
		if err := tx.Where("route_id = ?", route.ID).Delete(&models.RouteMiddleware{}).Error; err != nil {
			return err
		}

		if len(route.Middlewares) > 0 {
			routeMiddlewares := make([]models.RouteMiddleware, len(route.Middlewares))
			for i, name := range route.Middlewares {
				routeMiddlewares[i] = models.RouteMiddleware{
					RouteID:        route.ID,
					MiddlewareName: name,
					ExecutionOrder: i,
				}
			}
			if err := tx.Create(&routeMiddlewares).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete deletes a route and all its associations (cascade)
func (r *RouteRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Route{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("route not found: %d", id)
	}
	return nil
}

// DeleteByName deletes a route by name
func (r *RouteRepository) DeleteByName(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&models.Route{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("route not found: %s", name)
	}
	return nil
}

// Exists checks if a route exists by name
func (r *RouteRepository) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Route{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

// FindByPath finds routes matching a specific path
func (r *RouteRepository) FindByPath(ctx context.Context, path string) ([]models.Route, error) {
	var routes []models.Route

	err := r.db.WithContext(ctx).
		Where("path = ?", path).
		Preload("Backends").
		Preload("Maintenance").
		Preload("TLSCertificates").
		Preload("HealthCheck").
		Preload("Security").
		Preload("RouteMiddlewares", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_middlewares.execution_order ASC")
		}).
		Order("priority DESC").
		Find(&routes).Error

	if err != nil {
		return nil, err
	}

	return routes, nil
}
