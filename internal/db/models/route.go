package models

import (
	"time"

	"gorm.io/gorm"
)

type Route struct {
	ID             uint        `gorm:"primaryKey" json:"id" yaml:"id"`
	Name           string      `gorm:"uniqueIndex;not null;size:255" json:"name" yaml:"name"`
	Path           string      `gorm:"not null;size:500" json:"path" yaml:"path"`
	Rewrite        *string     `gorm:"size:500" json:"rewrite,omitempty" yaml:"rewrite,omitempty"`
	Priority       int         `gorm:"default:0;index:idx_priority,priority:desc" json:"priority,omitempty" yaml:"priority,omitempty"`
	Enabled        bool        `gorm:"default:true;index" json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Methods        StringArray `gorm:"type:text[]" json:"methods,omitempty" yaml:"methods,omitempty"`
	Hosts          StringArray `gorm:"type:text[]" json:"hosts,omitempty" yaml:"hosts,omitempty"`
	Target         *string     `gorm:"size:500" json:"target,omitempty" yaml:"target,omitempty"`
	DisableMetrics bool        `gorm:"default:false" json:"disableMetrics,omitempty" yaml:"disableMetrics,omitempty"`
	CreatedAt      time.Time   `gorm:"column:created_at" json:"-" yaml:"-"`
	UpdatedAt      time.Time   `gorm:"column:updated_at" json:"-" yaml:"-"`

	// Associations
	Backends         []Backend         `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"backends,omitempty" yaml:"backends,omitempty"`
	Maintenance      *Maintenance      `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"maintenance,omitempty" yaml:"maintenance,omitempty"`
	TLSCertificates  []TLSCertificate  `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"-" yaml:"-"`
	TLS              *TLSWrapper       `gorm:"-" json:"tls,omitempty" yaml:"tls,omitempty"`
	HealthCheck      *HealthCheck      `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	Security         *Security         `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"security,omitempty" yaml:"security,omitempty"`
	RouteMiddlewares []RouteMiddleware `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"-" yaml:"-"`
	Middlewares      []string          `gorm:"-" json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
}

// TLSWrapper wraps TLS certificates for JSON/YAML output
type TLSWrapper struct {
	Certificates []TLSCertificate `json:"certificates,omitempty" yaml:"certificates,omitempty"`
}

func (Route) TableName() string {
	return "routes"
}

// AfterFind hook to populate virtual fields
func (r *Route) AfterFind(tx *gorm.DB) error {
	// Populate TLS wrapper from TLSCertificates
	if len(r.TLSCertificates) > 0 {
		r.TLS = &TLSWrapper{Certificates: r.TLSCertificates}
	}

	// Populate Middlewares array from RouteMiddlewares
	if len(r.RouteMiddlewares) > 0 {
		r.Middlewares = make([]string, len(r.RouteMiddlewares))
		for i, rm := range r.RouteMiddlewares {
			r.Middlewares[i] = rm.MiddlewareName
		}
	}

	return nil
}

// BeforeSave hook to handle virtual fields
func (r *Route) BeforeSave(tx *gorm.DB) error {
	// Convert TLS wrapper to TLSCertificates
	if r.TLS != nil && len(r.TLS.Certificates) > 0 {
		r.TLSCertificates = r.TLS.Certificates
	}

	// Convert Middlewares to RouteMiddlewares
	if len(r.Middlewares) > 0 {
		r.RouteMiddlewares = make([]RouteMiddleware, len(r.Middlewares))
		for i, name := range r.Middlewares {
			r.RouteMiddlewares[i] = RouteMiddleware{
				MiddlewareName: name,
				ExecutionOrder: i,
			}
		}
	}

	return nil
}

type Backend struct {
	ID        uint      `gorm:"primaryKey" json:"-" yaml:"-"`
	RouteID   uint      `gorm:"not null;index" json:"-" yaml:"-"`
	Endpoint  string    `gorm:"not null;size:500" json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Weight    int       `gorm:"default:1" json:"weight,omitempty" yaml:"weight,omitempty"`
	Exclusive bool      `gorm:"default:false" json:"exclusive,omitempty" yaml:"exclusive,omitempty"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-" yaml:"-"`
}

type Maintenance struct {
	ID         uint      `gorm:"primaryKey" json:"-" yaml:"-"`
	RouteID    uint      `gorm:"uniqueIndex;not null" json:"-" yaml:"-"`
	Enabled    bool      `gorm:"default:false" json:"enabled,omitempty" yaml:"enabled,omitempty"`
	StatusCode int       `gorm:"default:503" json:"statusCode,omitempty" yaml:"statusCode,omitempty"`
	Message    string    `gorm:"default:'Service temporarily unavailable'" json:"message,omitempty" yaml:"message,omitempty"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"-" yaml:"-"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"-" yaml:"-"`
}

type TLSCertificate struct {
	ID        uint      `gorm:"primaryKey" json:"-" yaml:"-"`
	RouteID   uint      `gorm:"not null;index" json:"-" yaml:"-"`
	Cert      string    `gorm:"type:text;not null" json:"cert" yaml:"cert"`
	Key       string    `gorm:"type:text;not null" json:"key" yaml:"key"`
	CreatedAt time.Time `gorm:"column:created_at" json:"-" yaml:"-"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"-" yaml:"-"`
}

type HealthCheck struct {
	ID              uint      `gorm:"primaryKey" json:"-" yaml:"-"`
	RouteID         uint      `gorm:"uniqueIndex;not null" json:"-" yaml:"-"`
	Path            *string   `gorm:"size:500" json:"path,omitempty" yaml:"path,omitempty"`
	Interval        *string   `gorm:"size:50" json:"interval,omitempty" yaml:"interval,omitempty"`
	Timeout         *string   `gorm:"size:50" json:"timeout,omitempty" yaml:"timeout,omitempty"`
	HealthyStatuses IntArray  `gorm:"type:integer[]" json:"healthyStatuses,omitempty" yaml:"healthyStatuses,omitempty"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"-" yaml:"-"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"-" yaml:"-"`
}

type Security struct {
	ID                      uint         `gorm:"primaryKey" json:"-" yaml:"-"`
	RouteID                 uint         `gorm:"uniqueIndex;not null" json:"-" yaml:"-"`
	ForwardHostHeaders      bool         `gorm:"default:true" json:"forwardHostHeaders" yaml:"forwardHostHeaders"`
	EnableExploitProtection bool         `gorm:"default:false" json:"enableExploitProtection" yaml:"enableExploitProtection"`
	TLS                     *SecurityTLS `gorm:"embedded;embeddedPrefix:tls_" json:"tls,omitempty" yaml:"tls,omitempty"`
	CreatedAt               time.Time    `gorm:"column:created_at" json:"-" yaml:"-"`
	UpdatedAt               time.Time    `gorm:"column:updated_at" json:"-" yaml:"-"`
}

type SecurityTLS struct {
	InsecureSkipVerify bool    `gorm:"column:insecure_skip_verify;default:false" json:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty"`
	RootCAs            *string `gorm:"column:root_cas;type:text" json:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	ClientCert         *string `gorm:"column:client_cert;type:text" json:"clientCert,omitempty" yaml:"clientCert,omitempty"`
	ClientKey          *string `gorm:"column:client_key;type:text" json:"clientKey,omitempty" yaml:"clientKey,omitempty"`
}

type RouteMiddleware struct {
	ID             uint      `gorm:"primaryKey" json:"-" yaml:"-"`
	RouteID        uint      `gorm:"not null;index:idx_route_middleware_order" json:"-" yaml:"-"`
	MiddlewareName string    `gorm:"not null;size:255;uniqueIndex:idx_route_middleware_unique" json:"-" yaml:"-"`
	ExecutionOrder int       `gorm:"default:0;index:idx_route_middleware_order" json:"-" yaml:"-"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"-" yaml:"-"`

	// Associations
	Route      Route      `gorm:"foreignKey:RouteID;constraint:OnDelete:CASCADE" json:"-" yaml:"-"`
	Middleware Middleware `gorm:"foreignKey:MiddlewareName;references:Name" json:"-" yaml:"-"`
}
