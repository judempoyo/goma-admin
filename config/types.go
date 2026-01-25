package config

import (
	"gorm.io/gorm"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
	Cors     CorsConfig
	JWT      JWTConfig
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
	Secret string
}

type LogConfig struct {
	Level string
}
