package config

import (
	"fmt"
	"strings"
	"time"

	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/goma-admin/util"
	"github.com/jkaninda/okapi"
	"github.com/jkaninda/okapi/okapicli"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(app *okapi.Okapi, cli *okapicli.CLI) (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()
	// Parse flags
	if err := cli.ParseFlags(); err != nil {
		return nil, err
	}
	port := cli.GetInt("port")
	cfg := &Config{
		Database: DatabaseConfig{
			dbHost:     goutils.Env("GOMA_DB_HOST", "localhost"),
			dbUser:     goutils.Env("GOMA_DB_USER", "goma-admin"),
			dbPassword: goutils.Env("GOMA_DB_PASSWORD", "goma-admin"),
			dbName:     goutils.Env("GOMA_DB_NAME", "goma-admin"),
			dbPort:     goutils.EnvInt("GOMA_DB_PORT", 5432),
			dbSslMode:  goutils.Env("GOMA_DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			URL: goutils.Env("GOMA_REDIS_URL", "redis://localhost:6379/0"),
		},
		Server: ServerConfig{
			enableDocs:  goutils.EnvBool("GOMA_ENABLE_DOCS", true),
			Port:        goutils.EnvInt("GOMA_PORT", port),
			Environment: goutils.Env("GOMA_ENVIRONMENT", "development"),
		},
		Cors: CorsConfig{
			AllowedOrigins: strings.Split(goutils.Env("GOMA_CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		},
		JWT: JWTConfig{
			Secret:          goutils.Env("GOMA_JWT_SECRET", "default-secret-key"),
			Issuer:          goutils.Env("GOMA_JWT_ISSUER", "goma-admin"),
			Audience:        goutils.Env("GOMA_JWT_AUDIENCE", "goma-admin"),
			AccessTokenTTL:  envDuration("GOMA_JWT_ACCESS_TOKEN_TTL", "15m"),
			RefreshTokenTTL: envDuration("GOMA_JWT_REFRESH_TOKEN_TTL", "168h"),
		},
		Auth: AuthConfig{
			AllowFirstAdmin:      goutils.EnvBool("GOMA_AUTH_ALLOW_FIRST_ADMIN", false),
			RequireVerifiedEmail: goutils.EnvBool("GOMA_AUTH_REQUIRE_VERIFIED_EMAIL", false),
			PasswordPolicy: PasswordPolicyConfig{
				MinLength:      goutils.EnvInt("GOMA_PASSWORD_MIN_LENGTH", 12),
				MaxLength:      goutils.EnvInt("GOMA_PASSWORD_MAX_LENGTH", 128),
				RequireUpper:   goutils.EnvBool("GOMA_PASSWORD_REQUIRE_UPPER", true),
				RequireLower:   goutils.EnvBool("GOMA_PASSWORD_REQUIRE_LOWER", true),
				RequireNumber:  goutils.EnvBool("GOMA_PASSWORD_REQUIRE_NUMBER", true),
				RequireSpecial: goutils.EnvBool("GOMA_PASSWORD_REQUIRE_SPECIAL", true),
			},
			OAuth: OAuthConfig{
				Google: OAuthProviderConfig{
					ClientID:     goutils.Env("GOMA_OAUTH_GOOGLE_CLIENT_ID", ""),
					ClientSecret: goutils.Env("GOMA_OAUTH_GOOGLE_CLIENT_SECRET", ""),
					RedirectURL:  goutils.Env("GOMA_OAUTH_GOOGLE_REDIRECT_URL", ""),
					AuthURL:      goutils.Env("GOMA_OAUTH_GOOGLE_AUTH_URL", ""),
					TokenURL:     goutils.Env("GOMA_OAUTH_GOOGLE_TOKEN_URL", ""),
					UserInfoURL:  goutils.Env("GOMA_OAUTH_GOOGLE_USERINFO_URL", ""),
				},
				GitHub: OAuthProviderConfig{
					ClientID:     goutils.Env("GOMA_OAUTH_GITHUB_CLIENT_ID", ""),
					ClientSecret: goutils.Env("GOMA_OAUTH_GITHUB_CLIENT_SECRET", ""),
					RedirectURL:  goutils.Env("GOMA_OAUTH_GITHUB_REDIRECT_URL", ""),
					AuthURL:      goutils.Env("GOMA_OAUTH_GITHUB_AUTH_URL", ""),
					TokenURL:     goutils.Env("GOMA_OAUTH_GITHUB_TOKEN_URL", ""),
					UserInfoURL:  goutils.Env("GOMA_OAUTH_GITHUB_USERINFO_URL", ""),
					EmailsURL:    goutils.Env("GOMA_OAUTH_GITHUB_EMAILS_URL", ""),
				},
			},
		},
		Log: LogConfig{
			Level: goutils.Env("GOMA_LOG_LEVEL", "info"),
		},
	}
	if err := cfg.initialize(app); err != nil {
		return nil, err
	}
	return cfg, nil

}
func (c *Config) validate() error {
	if c.Redis.URL == "" {
		return fmt.Errorf("GOMA_REDIS_URL is required")
	}
	if c.Server.Port == 0 {
		return fmt.Errorf("GOMA_PORT is required")
	}
	if c.JWT.Secret == "" {
		return fmt.Errorf("GOMA_JWT_SECRET is required")
	}
	return nil
}

func (c *Config) initialize(app *okapi.Okapi) error {
	if err := c.validate(); err != nil {
		return err
	}
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", c.Database.dbHost, c.Database.dbUser, c.Database.dbPassword, c.Database.dbName, c.Database.dbPort, c.Database.dbSslMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	c.Database.DB = db
	// Init Doc
	if c.Server.enableDocs {
		app.WithOpenAPIDocs(okapi.OpenAPI{
			Title:   util.AppName,
			Version: util.AppVersion,
		})
	}
	app.WithPort(c.Server.Port)
	return nil
}

func envDuration(key, fallback string) time.Duration {
	raw := goutils.Env(key, fallback)
	parsed, err := time.ParseDuration(raw)
	if err == nil {
		return parsed
	}
	parsed, err = time.ParseDuration(fallback)
	if err == nil {
		return parsed
	}
	return 15 * time.Minute
}
