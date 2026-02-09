package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/internal/db/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InstanceRepository struct {
	db *gorm.DB
}

func NewInstanceRepository(db *gorm.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// Create creates a new instance
func (r *InstanceRepository) Create(ctx context.Context, instance *models.Instance) error {
	return r.db.WithContext(ctx).Create(instance).Error
}

// GetByID retrieves an instance by ID with routes
func (r *InstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Instance, error) {
	var instance models.Instance

	err := r.db.WithContext(ctx).
		Preload("Routes").
		Preload("InstanceRoutes").
		First(&instance, "id = ?", id).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("instance not found: %s", id)
		}
		return nil, err
	}

	return &instance, nil
}

// GetByName retrieves an instance by name with routes
func (r *InstanceRepository) GetByName(ctx context.Context, name string) (*models.Instance, error) {
	var instance models.Instance

	err := r.db.WithContext(ctx).
		Preload("Routes").
		Preload("InstanceRoutes").
		Where("name = ?", name).
		First(&instance).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("instance not found: %s", name)
		}
		return nil, err
	}

	return &instance, nil
}

// List retrieves all instances
func (r *InstanceRepository) List(ctx context.Context) ([]models.Instance, error) {
	var instances []models.Instance

	err := r.db.WithContext(ctx).
		Preload("Routes").
		Order("name ASC").
		Find(&instances).Error

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// ListByEnvironment retrieves instances by environment
func (r *InstanceRepository) ListByEnvironment(ctx context.Context, environment string) ([]models.Instance, error) {
	var instances []models.Instance

	err := r.db.WithContext(ctx).
		Where("environment = ?", environment).
		Preload("Routes").
		Order("name ASC").
		Find(&instances).Error

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// ListByStatus retrieves instances by status
func (r *InstanceRepository) ListByStatus(ctx context.Context, status string) ([]models.Instance, error) {
	var instances []models.Instance

	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Preload("Routes").
		Order("name ASC").
		Find(&instances).Error

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// ListActive retrieves all active instances
func (r *InstanceRepository) ListActive(ctx context.Context) ([]models.Instance, error) {
	var instances []models.Instance

	err := r.db.WithContext(ctx).
		Where("enabled = ? AND status = ?", true, "active").
		Preload("Routes").
		Order("name ASC").
		Find(&instances).Error

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// Update updates an instance
func (r *InstanceRepository) Update(ctx context.Context, instance *models.Instance) error {
	return r.db.WithContext(ctx).
		Model(instance).
		Updates(map[string]interface{}{
			"name":             instance.Name,
			"environment":      instance.Environment,
			"description":      instance.Description,
			"endpoint":         instance.Endpoint,
			"metrics_endpoint": instance.MetricsEndpoint,
			"health_endpoint":  instance.HealthEndpoint,
			"version":          instance.Version,
			"region":           instance.Region,
			"tags":             instance.Tags,
			"status":           instance.Status,
			"enabled":          instance.Enabled,
			"metadata":         instance.Metadata,
			"last_seen":        instance.LastSeen,
		}).Error
}

// UpdateStatus updates instance status and last seen time
func (r *InstanceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    status,
			"last_seen": now,
		}).Error
}

// UpdateLastSeen updates the last seen timestamp
func (r *InstanceRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Where("id = ?", id).
		Update("last_seen", now).Error
}

// Delete deletes an instance
func (r *InstanceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&models.Instance{}, "id = ?", id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("instance not found: %s", id)
	}

	return nil
}

// Exists checks if an instance exists by name
func (r *InstanceRepository) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Where("name = ?", name).
		Count(&count).Error

	return count > 0, err
}

// ----- Instance-Route Association Methods -----

// AttachRoute attaches a route to an instance
func (r *InstanceRepository) AttachRoute(ctx context.Context, instanceID uuid.UUID, routeID uint, options *models.InstanceRoute) error {
	instanceRoute := &models.InstanceRoute{
		InstanceID: instanceID,
		RouteID:    routeID,
		Enabled:    true,
	}

	if options != nil {
		if options.Priority != nil {
			instanceRoute.Priority = options.Priority
		}
		instanceRoute.Enabled = options.Enabled
		instanceRoute.DeployedBy = options.DeployedBy
		instanceRoute.ConfigVersion = options.ConfigVersion
		instanceRoute.Metadata = options.Metadata
	}

	now := time.Now()
	instanceRoute.DeployedAt = &now

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "instance_id"}, {Name: "route_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"enabled", "priority", "deployed_at", "deployed_by",
				"config_version", "metadata", "updated_at",
			}),
		}).
		Create(instanceRoute).Error
}

// DetachRoute removes a route from an instance
func (r *InstanceRepository) DetachRoute(ctx context.Context, instanceID uuid.UUID, routeID uint) error {
	result := r.db.WithContext(ctx).
		Where("instance_id = ? AND route_id = ?", instanceID, routeID).
		Delete(&models.InstanceRoute{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("route not attached to instance")
	}

	return nil
}

// AttachRoutes attaches multiple routes to an instance
func (r *InstanceRepository) AttachRoutes(ctx context.Context, instanceID uuid.UUID, routeIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for _, routeID := range routeIDs {
			instanceRoute := &models.InstanceRoute{
				InstanceID: instanceID,
				RouteID:    routeID,
				Enabled:    true,
				DeployedAt: &now,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "instance_id"}, {Name: "route_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"enabled", "deployed_at", "updated_at"}),
			}).Create(instanceRoute).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DetachRoutes removes multiple routes from an instance
func (r *InstanceRepository) DetachRoutes(ctx context.Context, instanceID uuid.UUID, routeIDs []uint) error {
	return r.db.WithContext(ctx).
		Where("instance_id = ? AND route_id IN ?", instanceID, routeIDs).
		Delete(&models.InstanceRoute{}).Error
}

// SyncRoutes replaces all routes for an instance
func (r *InstanceRepository) SyncRoutes(ctx context.Context, instanceID uuid.UUID, routeIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing routes
		if err := tx.Where("instance_id = ?", instanceID).Delete(&models.InstanceRoute{}).Error; err != nil {
			return err
		}

		// Add new routes
		if len(routeIDs) == 0 {
			return nil
		}

		now := time.Now()
		instanceRoutes := make([]models.InstanceRoute, len(routeIDs))
		for i, routeID := range routeIDs {
			instanceRoutes[i] = models.InstanceRoute{
				InstanceID: instanceID,
				RouteID:    routeID,
				Enabled:    true,
				DeployedAt: &now,
			}
		}

		return tx.Create(&instanceRoutes).Error
	})
}

// GetRoutesByInstance retrieves all routes for a specific instance
func (r *InstanceRepository) GetRoutesByInstance(ctx context.Context, instanceID uuid.UUID) ([]models.Route, error) {
	var routes []models.Route

	err := r.db.WithContext(ctx).
		Joins("INNER JOIN instance_routes ON instance_routes.route_id = routes.id").
		Where("instance_routes.instance_id = ?", instanceID).
		Preload("Backends").
		Preload("Maintenance").
		Preload("TLSCertificates").
		Preload("HealthCheck").
		Preload("Security").
		Preload("RouteMiddlewares", func(db *gorm.DB) *gorm.DB {
			return db.Order("route_middlewares.execution_order ASC")
		}).
		Order("routes.priority DESC, routes.name ASC").
		Find(&routes).Error

	if err != nil {
		return nil, err
	}

	return routes, nil
}

// GetInstancesByRoute retrieves all instances that have a specific route
func (r *InstanceRepository) GetInstancesByRoute(ctx context.Context, routeID uint) ([]models.Instance, error) {
	var instances []models.Instance

	err := r.db.WithContext(ctx).
		Joins("INNER JOIN instance_routes ON instance_routes.instance_id = instances.id").
		Where("instance_routes.route_id = ?", routeID).
		Order("instances.name ASC").
		Find(&instances).Error

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// GetInstanceRoute retrieves the instance-route relationship
func (r *InstanceRepository) GetInstanceRoute(ctx context.Context, instanceID uuid.UUID, routeID uint) (*models.InstanceRoute, error) {
	var instanceRoute models.InstanceRoute

	err := r.db.WithContext(ctx).
		Where("instance_id = ? AND route_id = ?", instanceID, routeID).
		First(&instanceRoute).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("route not attached to instance")
		}
		return nil, err
	}

	return &instanceRoute, nil
}

// UpdateInstanceRoute updates instance-specific route configuration
func (r *InstanceRepository) UpdateInstanceRoute(ctx context.Context, instanceRoute *models.InstanceRoute) error {
	return r.db.WithContext(ctx).
		Model(instanceRoute).
		Where("instance_id = ? AND route_id = ?", instanceRoute.InstanceID, instanceRoute.RouteID).
		Updates(map[string]interface{}{
			"enabled":        instanceRoute.Enabled,
			"priority":       instanceRoute.Priority,
			"deployed_by":    instanceRoute.DeployedBy,
			"config_version": instanceRoute.ConfigVersion,
			"metadata":       instanceRoute.Metadata,
		}).Error
}

// GetHealthyInstances retrieves all healthy instances
func (r *InstanceRepository) GetHealthyInstances(ctx context.Context) ([]models.Instance, error) {
	var instances []models.Instance

	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	err := r.db.WithContext(ctx).
		Where("enabled = ? AND status = ? AND last_seen >= ?", true, "active", fiveMinutesAgo).
		Preload("Routes").
		Order("name ASC").
		Find(&instances).Error

	if err != nil {
		return nil, err
	}

	return instances, nil
}

// Count returns the total number of instances
func (r *InstanceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Instance{}).Count(&count).Error
	return count, err
}

// CountByEnvironment returns the number of instances per environment
func (r *InstanceRepository) CountByEnvironment(ctx context.Context, environment string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Where("environment = ?", environment).
		Count(&count).Error
	return count, err
}

// GetInstanceStats returns statistics about instances
func (r *InstanceRepository) GetInstanceStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	total, err := r.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count by environment
	var envCounts []struct {
		Environment string
		Count       int64
	}

	err = r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Select("environment, COUNT(*) as count").
		Group("environment").
		Order("count DESC").
		Scan(&envCounts).Error

	if err != nil {
		return nil, err
	}
	stats["byEnvironment"] = envCounts

	// Count by status
	var statusCounts []struct {
		Status string
		Count  int64
	}

	err = r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Order("count DESC").
		Scan(&statusCounts).Error

	if err != nil {
		return nil, err
	}
	stats["byStatus"] = statusCounts

	// Active instances
	activeCount := r.db.WithContext(ctx).
		Model(&models.Instance{}).
		Where("enabled = ? AND status = ?", true, "active").
		Count(&total)
	stats["active"] = activeCount

	return stats, nil
}
