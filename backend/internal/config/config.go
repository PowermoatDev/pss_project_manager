package config

import (
	"fmt"
	"os"
)

type Config struct {
	Host          string
	Port          string
	DatabaseURL   string
	AutoMigrate   bool
	UploadDir     string
	AllowedOrigin string
}

func Load() Config {
	return Config{
		Host:          value("APP_HOST", "0.0.0.0"),
		Port:          value("APP_PORT", "8080"),
		DatabaseURL:   value("DATABASE_URL", defaultDatabaseURL()),
		AutoMigrate:   value("AUTO_MIGRATE", "true") != "false",
		UploadDir:     value("UPLOAD_DIR", "uploads"),
		AllowedOrigin: value("ALLOWED_ORIGIN", "*"),
	}
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func value(key string, fallback string) string {
	if found := os.Getenv(key); found != "" {
		return found
	}
	return fallback
}

func defaultDatabaseURL() string {
	return "sqlserver://sa:YourStrong!Passw0rd@localhost:1433?database=PrintSecWarRoom&encrypt=disable"
}
