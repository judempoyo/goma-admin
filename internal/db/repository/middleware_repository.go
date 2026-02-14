package repository

import (
	"context"
	"fmt"

	"github.com/jkaninda/goma-admin/internal/db/models"
	"gorm.io/gorm"
)

type MiddlewareRepository struct {
	db *gorm.DB
}

func NewMiddlewareRepository(db *gorm.DB) *MiddlewareRepository {
	return &MiddlewareRepository{db: db}
}

// Create creates a new middleware
func (r *MiddlewareRepository) Create(ctx context.Context, middleware *models.Middleware) error {
	if err := r.db.WithContext(ctx).Create(middleware).Error; err != nil {
		return fmt.Errorf("failed to create middleware: %w", err)
	}
	return nil
}

// CreateBatch creates multiple middlewares in a single transaction
func (r *MiddlewareRepository) CreateBatch(ctx context.Context, middlewares []models.Middleware) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range middlewares {
			if err := tx.Create(&middlewares[i]).Error; err != nil {
				return fmt.Errorf("failed to create middleware %s: %w", middlewares[i].Name, err)
			}
		}
		return nil
	})
}

// GetByID retrieves a middleware by ID
func (r *MiddlewareRepository) GetByID(ctx context.Context, id uint) (*models.Middleware, error) {
	var middleware models.Middleware

	err := r.db.WithContext(ctx).First(&middleware, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("middleware not found: %d", id)
		}
		return nil, err
	}

	return &middleware, nil
}

// GetByName retrieves a middleware by name
func (r *MiddlewareRepository) GetByName(ctx context.Context, name string) (*models.Middleware, error) {
	var middleware models.Middleware

	err := r.db.WithContext(ctx).Where("name = ?", name).First(&middleware).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("middleware not found: %s", name)
		}
		return nil, err
	}

	return &middleware, nil
}

// GetByNames retrieves multiple middlewares by their names
func (r *MiddlewareRepository) GetByNames(ctx context.Context, names []string) ([]models.Middleware, error) {
	if len(names) == 0 {
		return []models.Middleware{}, nil
	}

	var middlewares []models.Middleware

	err := r.db.WithContext(ctx).
		Where("name IN ?", names).
		Find(&middlewares).Error

	if err != nil {
		return nil, err
	}

	return middlewares, nil
}

// List retrieves all middlewares
func (r *MiddlewareRepository) List(ctx context.Context) ([]models.Middleware, error) {
	var middlewares []models.Middleware

	err := r.db.WithContext(ctx).
		Order("name ASC").
		Find(&middlewares).Error

	if err != nil {
		return nil, err
	}

	return middlewares, nil
}

// ListByType retrieves middlewares of a specific type
func (r *MiddlewareRepository) ListByType(ctx context.Context, middlewareType string) ([]models.Middleware, error) {
	var middlewares []models.Middleware

	err := r.db.WithContext(ctx).
		Where("type = ?", middlewareType).
		Order("name ASC").
		Find(&middlewares).Error

	if err != nil {
		return nil, err
	}

	return middlewares, nil
}

// ListByTypes retrieves middlewares of multiple types
func (r *MiddlewareRepository) ListByTypes(ctx context.Context, types []string) ([]models.Middleware, error) {
	if len(types) == 0 {
		return []models.Middleware{}, nil
	}

	var middlewares []models.Middleware

	err := r.db.WithContext(ctx).
		Where("type IN ?", types).
		Order("type ASC, name ASC").
		Find(&middlewares).Error

	if err != nil {
		return nil, err
	}

	return middlewares, nil
}

// ListWithPagination retrieves middlewares with pagination
func (r *MiddlewareRepository) ListWithPagination(ctx context.Context, page, pageSize int) ([]models.Middleware, int64, error) {
	var middlewares []models.Middleware
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&models.Middleware{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Fetch paginated records
	err := r.db.WithContext(ctx).
		Order("name ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&middlewares).Error

	if err != nil {
		return nil, 0, err
	}

	return middlewares, total, nil
}

// Search searches middlewares by name or type (case-insensitive)
func (r *MiddlewareRepository) Search(ctx context.Context, query string) ([]models.Middleware, error) {
	var middlewares []models.Middleware

	searchPattern := "%" + query + "%"

	err := r.db.WithContext(ctx).
		Where("name ILIKE ? OR type ILIKE ?", searchPattern, searchPattern).
		Order("name ASC").
		Find(&middlewares).Error

	if err != nil {
		return nil, err
	}

	return middlewares, nil
}

// Update updates a middleware
func (r *MiddlewareRepository) Update(ctx context.Context, middleware *models.Middleware) error {
	result := r.db.WithContext(ctx).Model(middleware).Updates(map[string]interface{}{
		"type":  middleware.Type,
		"paths": middleware.Paths,
		"rule":  middleware.Rule,
	})

	if result.Error != nil {
		return fmt.Errorf("failed to update middleware: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("middleware not found: %s", middleware.Name)
	}

	return nil
}

// UpdateByName updates a middleware by name
func (r *MiddlewareRepository) UpdateByName(ctx context.Context, name string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&models.Middleware{}).
		Where("name = ?", name).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update middleware: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("middleware not found: %s", name)
	}

	return nil
}

// Delete deletes a middleware by ID
func (r *MiddlewareRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Middleware{}, id)

	if result.Error != nil {
		return fmt.Errorf("failed to delete middleware: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("middleware not found: %d", id)
	}

	return nil
}

// DeleteByName deletes a middleware by name
func (r *MiddlewareRepository) DeleteByName(ctx context.Context, name string) error {
	result := r.db.WithContext(ctx).Where("name = ?", name).Delete(&models.Middleware{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete middleware: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("middleware not found: %s", name)
	}

	return nil
}

// DeleteBatch deletes multiple middlewares by their names
func (r *MiddlewareRepository) DeleteBatch(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("name IN ?", names).Delete(&models.Middleware{})

		if result.Error != nil {
			return fmt.Errorf("failed to delete middlewares: %w", result.Error)
		}

		return nil
	})
}

// Exists checks if a middleware exists by name
func (r *MiddlewareRepository) Exists(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Middleware{}).
		Where("name = ?", name).
		Count(&count).Error

	return count > 0, err
}

// ExistsByNames checks which middlewares exist from a list of names
func (r *MiddlewareRepository) ExistsByNames(ctx context.Context, names []string) (map[string]bool, error) {
	if len(names) == 0 {
		return make(map[string]bool), nil
	}

	var existingNames []string
	err := r.db.WithContext(ctx).
		Model(&models.Middleware{}).
		Where("name IN ?", names).
		Pluck("name", &existingNames).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, name := range names {
		result[name] = false
	}
	for _, name := range existingNames {
		result[name] = true
	}

	return result, nil
}

// Count returns the total number of middlewares
func (r *MiddlewareRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Middleware{}).Count(&count).Error
	return count, err
}

// CountByType returns the number of middlewares for a specific type
func (r *MiddlewareRepository) CountByType(ctx context.Context, middlewareType string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Middleware{}).
		Where("type = ?", middlewareType).
		Count(&count).Error
	return count, err
}

// GetRoutesByMiddleware retrieves all routes using a specific middleware
func (r *MiddlewareRepository) GetRoutesByMiddleware(ctx context.Context, middlewareName string) ([]models.Route, error) {
	var routes []models.Route

	err := r.db.WithContext(ctx).
		Joins("INNER JOIN route_middlewares ON route_middlewares.route_id = routes.id").
		Where("route_middlewares.middleware_name = ?", middlewareName).
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

// GetUsageCount returns the number of routes using a specific middleware
func (r *MiddlewareRepository) GetUsageCount(ctx context.Context, middlewareName string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.RouteMiddleware{}).
		Where("middleware_name = ?", middlewareName).
		Count(&count).Error

	return count, err
}

// IsMiddlewareInUse checks if a middleware is currently being used by any route
func (r *MiddlewareRepository) IsMiddlewareInUse(ctx context.Context, middlewareName string) (bool, error) {
	count, err := r.GetUsageCount(ctx, middlewareName)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetMiddlewareStats returns statistics about middleware usage
func (r *MiddlewareRepository) GetMiddlewareStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	total, err := r.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count by type
	var typeCounts []struct {
		Type  string
		Count int64
	}

	err = r.db.WithContext(ctx).
		Model(&models.Middleware{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Order("count DESC").
		Scan(&typeCounts).Error

	if err != nil {
		return nil, err
	}
	stats["byType"] = typeCounts

	// Most used middlewares
	var mostUsed []struct {
		Name  string
		Type  string
		Count int64
	}

	err = r.db.WithContext(ctx).
		Table("middlewares").
		Select("middlewares.name, middlewares.type, COUNT(route_middlewares.id) as count").
		Joins("LEFT JOIN route_middlewares ON route_middlewares.middleware_name = middlewares.name").
		Group("middlewares.name, middlewares.type").
		Order("count DESC").
		Limit(10).
		Scan(&mostUsed).Error

	if err != nil {
		return nil, err
	}
	stats["mostUsed"] = mostUsed

	return stats, nil
}
