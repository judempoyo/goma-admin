package models

import (
	"time"

	"gorm.io/gorm"
)

type Middleware struct {
	ID        uint        `gorm:"primaryKey" json:"-" yaml:"-"`
	Name      string      `gorm:"uniqueIndex;not null;size:255" json:"name" yaml:"name"`
	Type      string      `gorm:"not null;size:100;index" json:"type" yaml:"type"`
	Paths     StringArray `gorm:"type:text[]" json:"paths,omitempty" yaml:"paths,omitempty"`
	Rule      JSONB       `gorm:"type:jsonb" json:"rule,omitempty" yaml:"rule,omitempty"`
	CreatedAt time.Time   `gorm:"column:created_at" json:"-" yaml:"-"`
	UpdatedAt time.Time   `gorm:"column:updated_at" json:"-" yaml:"-"`

	// Associations
	RouteMiddlewares []RouteMiddleware `gorm:"foreignKey:MiddlewareName;references:Name;constraint:OnDelete:CASCADE" json:"-" yaml:"-"`
}

// TableName specifies the table name for the Middleware model
func (Middleware) TableName() string {
	return "middlewares"
}

// BeforeDelete hook to handle cascade deletion if needed
func (m *Middleware) BeforeDelete(tx *gorm.DB) error {
	return nil
}
