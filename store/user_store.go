package store

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/models"
	"gorm.io/gorm"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	Count(ctx context.Context) (int64, error)
	AddRoles(ctx context.Context, user *models.User, roles []models.Role) error
	ReplaceRoles(ctx context.Context, user *models.User, roles []models.Role) error
	List(ctx context.Context, limit, offset int) ([]models.User, error)
}

type GormUserStore struct {
	db *gorm.DB
}

func NewUserStore(db *gorm.DB) *GormUserStore {
	return &GormUserStore{db: db}
}

func (s *GormUserStore) Create(ctx context.Context, user *models.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

func (s *GormUserStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("Roles").
		Where("email = ?", strings.ToLower(strings.TrimSpace(email))).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GormUserStore) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Preload("Roles").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *GormUserStore) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *GormUserStore) AddRoles(ctx context.Context, user *models.User, roles []models.Role) error {
	return s.db.WithContext(ctx).Model(user).Association("Roles").Append(roles)
}

func (s *GormUserStore) ReplaceRoles(ctx context.Context, user *models.User, roles []models.Role) error {
	return s.db.WithContext(ctx).Model(user).Association("Roles").Replace(roles)
}

func (s *GormUserStore) List(ctx context.Context, limit, offset int) ([]models.User, error) {
	var users []models.User
	query := s.db.WithContext(ctx).Preload("Roles").Order("created_at desc")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
