package config

import (
	"time"

	"gorm.io/gorm"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	Cors     CorsConfig
	JWT      JWTConfig
	Auth     AuthConfig
	Log      LogConfig
}

type DatabaseConfig struct {
	DB         *gorm.DB
	url        string
	dbHost     string
	dbUser     string
	dbPassword string
	dbName     string
	dbPort     int
	dbSslMode  string
}

type RedisConfig struct {
	URL string
}

type ServerConfig struct {
	Port        int
	Environment string
	enableDocs  bool
}

type CorsConfig struct {
	AllowedOrigins []string
}

type JWTConfig struct {
	Secret          string
	Issuer          string
	Audience        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type LogConfig struct {
	Level string
}

type AuthConfig struct {
	AllowFirstAdmin      bool
	RequireVerifiedEmail bool
	PasswordPolicy       PasswordPolicyConfig
	OAuth                OAuthConfig
}

type PasswordPolicyConfig struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireNumber  bool
	RequireSpecial bool
}

type OAuthConfig struct {
	Google OAuthProviderConfig
	GitHub OAuthProviderConfig
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	EmailsURL    string
}
