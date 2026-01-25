package config

import (
	"fmt"
	"strings"

	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/goma-admin/util"
	"github.com/jkaninda/logger"
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
			dbHost:     goutils.Env("DB_HOST", "localhost"),
			dbUser:     goutils.Env("DB_USER", "goma-admin"),
			dbPassword: goutils.Env("DB_PASSWORD", "goma-admin"),
			dbName:     goutils.Env("DB_NAME", "goma-admin"),
			dbPort:     goutils.EnvInt("DB_PORT", 5432),
			dbSslMode:  goutils.Env("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			URL: goutils.Env("REDIS_URL", "redis://localhost:6379/0"),
		},
		Server: ServerConfig{
			enableDocs:  goutils.EnvBool("ENABLE_DOCS", true),
			Port:        goutils.EnvInt("PORT", port),
			Environment: goutils.Env("ENVIRONMENT", "development"),
		},
		Cors: CorsConfig{
			AllowedOrigins: strings.Split(goutils.Env("CORS_ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		},
		JWT: JWTConfig{
			Secret: goutils.Env("JWT_SECRET", "default-secret-key"),
		},
		Log: LogConfig{
			Level: goutils.Env("LOG_LEVEL", "info"),
		},
	}
	if err := cfg.initialize(app); err != nil {
		return nil, err
	}
	return cfg, nil

}
func (c *Config) validate() error {
	if c.Redis.URL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.Server.Port == 0 {
		return fmt.Errorf("PORT is required")
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
		logger.Fatal("Failed to connect to database", "error", err)
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
