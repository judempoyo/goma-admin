package store

import (
	"context"
	"strings"

	"github.com/jkaninda/goma-admin/models"
	"gorm.io/gorm"
)

type RoleStore interface {
	EnsureRoles(ctx context.Context, names []string) ([]models.Role, error)
	FindByNames(ctx context.Context, names []string) ([]models.Role, error)
	FindByName(ctx context.Context, name string) (*models.Role, error)
	Create(ctx context.Context, role *models.Role) error
	List(ctx context.Context, limit, offset int) ([]models.Role, error)
}

type GormRoleStore struct {
	db *gorm.DB
}

func NewRoleStore(db *gorm.DB) *GormRoleStore {
	return &GormRoleStore{db: db}
}

func (s *GormRoleStore) EnsureRoles(ctx context.Context, names []string) ([]models.Role, error) {
	roles := make([]models.Role, 0, len(names))
	for _, name := range names {
		normalized := strings.ToLower(strings.TrimSpace(name))
		if normalized == "" {
			continue
		}
		role := models.Role{Name: normalized}
		if err := s.db.WithContext(ctx).Where("name = ?", normalized).FirstOrCreate(&role).Error; err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (s *GormRoleStore) FindByNames(ctx context.Context, names []string) ([]models.Role, error) {
	if len(names) == 0 {
		return []models.Role{}, nil
	}
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		value := strings.ToLower(strings.TrimSpace(name))
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	var roles []models.Role
	if err := s.db.WithContext(ctx).Where("name IN ?", normalized).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *GormRoleStore) FindByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := s.db.WithContext(ctx).Where("name = ?", strings.ToLower(strings.TrimSpace(name))).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (s *GormRoleStore) Create(ctx context.Context, role *models.Role) error {
	return s.db.WithContext(ctx).Create(role).Error
}

func (s *GormRoleStore) List(ctx context.Context, limit, offset int) ([]models.Role, error) {
	var roles []models.Role
	query := s.db.WithContext(ctx).Order("name asc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
