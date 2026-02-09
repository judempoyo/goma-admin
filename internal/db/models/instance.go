package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Instance represents a Goma Gateway instance/environment
type Instance struct {
	ID              uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id" yaml:"id"`
	Name            string      `gorm:"uniqueIndex;not null;size:255" json:"name" yaml:"name"`
	Environment     string      `gorm:"size:100;index" json:"environment" yaml:"environment"` // dev, staging, prod, etc.
	Description     string      `gorm:"type:text" json:"description,omitempty" yaml:"description,omitempty"`
	Endpoint        string      `gorm:"not null;size:500" json:"endpoint" yaml:"endpoint"`
	MetricsEndpoint string      `gorm:"size:500" json:"metricsEndpoint,omitempty" yaml:"metricsEndpoint,omitempty"`
	HealthEndpoint  string      `gorm:"size:500" json:"healthEndpoint,omitempty" yaml:"healthEndpoint,omitempty"`
	Version         string      `gorm:"size:50" json:"version,omitempty" yaml:"version,omitempty"`
	Region          string      `gorm:"size:100" json:"region,omitempty" yaml:"region,omitempty"`
	Tags            StringArray `gorm:"type:text[]" json:"tags,omitempty" yaml:"tags,omitempty"`
	LastSeen        *time.Time  `gorm:"index" json:"lastSeen,omitempty" yaml:"lastSeen,omitempty"`
	Status          string      `gorm:"size:50;default:'unknown';index" json:"status" yaml:"status"` // active, inactive, unhealthy, unknown
	Enabled         bool        `gorm:"default:true;index" json:"enabled" yaml:"enabled"`
	Metadata        JSONB       `gorm:"type:jsonb" json:"metadata,omitempty" yaml:"metadata,omitempty"`
	CreatedAt       time.Time   `gorm:"column:created_at" json:"createdAt" yaml:"createdAt"`
	UpdatedAt       time.Time   `gorm:"column:updated_at" json:"updatedAt" yaml:"updatedAt"`

	// Associations
	InstanceRoutes []InstanceRoute `gorm:"foreignKey:InstanceID;constraint:OnDelete:CASCADE" json:"-" yaml:"-"`
	Routes         []Route         `gorm:"many2many:instance_routes;" json:"routes,omitempty" yaml:"routes,omitempty"`
}

// InstanceRoute represents the many-to-many relationship between instances and routes
// This allows for instance-specific route configurations
type InstanceRoute struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	InstanceID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_instance_route" json:"instanceId"`
	RouteID    uint      `gorm:"not null;uniqueIndex:idx_instance_route" json:"routeId"`

	// Instance-specific route overrides
	Enabled  bool `gorm:"default:true" json:"enabled"`     // Override route enabled status for this instance
	Priority *int `gorm:"index" json:"priority,omitempty"` // Override route priority for this instance

	// Deployment information
	DeployedAt    *time.Time `json:"deployedAt,omitempty"`
	DeployedBy    string     `gorm:"size:255" json:"deployedBy,omitempty"`
	ConfigVersion string     `gorm:"size:100" json:"configVersion,omitempty"` // Track config version deployed

	// Metadata for this specific instance-route relationship
	Metadata  JSONB     `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updatedAt"`

	// Associations
	Instance Instance `gorm:"foreignKey:InstanceID;constraint:OnDelete:CASCADE" json:"-"`
	Route    Route    `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName specifies the table name for the Instance model
func (Instance) TableName() string {
	return "instances"
}

// TableName specifies the table name for the InstanceRoute model
func (InstanceRoute) TableName() string {
	return "instance_routes"
}

// BeforeCreate hook to generate UUID and set defaults
func (i *Instance) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}

	if i.Status == "" {
		i.Status = "unknown"
	}

	return nil
}

// IsHealthy checks if the instance is healthy based on status and last seen
func (i *Instance) IsHealthy() bool {
	if i.Status != "active" || !i.Enabled {
		return false
	}

	if i.LastSeen == nil {
		return false
	}

	// Consider unhealthy if not seen in last 5 minutes
	return time.Since(*i.LastSeen) < 5*time.Minute
}

// UpdateStatus updates the instance status and last seen time
func (i *Instance) UpdateStatus(status string) {
	i.Status = status
	now := time.Now()
	i.LastSeen = &now
}

// InstanceStatus represents possible instance statuses
type InstanceStatus string

const (
	InstanceStatusActive    InstanceStatus = "active"
	InstanceStatusInactive  InstanceStatus = "inactive"
	InstanceStatusUnhealthy InstanceStatus = "unhealthy"
	InstanceStatusUnknown   InstanceStatus = "unknown"
)

// InstanceEnvironment represents common environment types
type InstanceEnvironment string

const (
	EnvironmentDevelopment InstanceEnvironment = "development"
	EnvironmentStaging     InstanceEnvironment = "staging"
	EnvironmentProduction  InstanceEnvironment = "production"
	EnvironmentTesting     InstanceEnvironment = "testing"
)
