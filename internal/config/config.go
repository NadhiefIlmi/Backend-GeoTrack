package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DBConfig
}

type AppConfig struct {
	Name              string   `mapstructure:"APP_NAME"`
	Port              int      `mapstructure:"APP_PORT"`
	Env               string   `mapstructure:"APP_ENV"`
	AllowedOriginsRaw string   `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AllowedOrigins    []string `mapstructure:"-"`
}

type DBConfig struct {
	URL      string `mapstructure:"DATABASE_URL"`
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`
	MaxConns int    `mapstructure:"DB_MAX_CONNS"`
	MinConns int    `mapstructure:"DB_MIN_CONNS"`
}

// LoadConfig reads configuration from file or environment variables
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Default values
	v.SetDefault("APP_NAME", "GeoTrack Backend")
	v.SetDefault("APP_PORT", 8080)
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:8081,http://localhost:5173")

	v.SetDefault("DATABASE_URL", "")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_NAME", "geotrack_db")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("DB_MAX_CONNS", 25)
	v.SetDefault("DB_MIN_CONNS", 5)

	if path != "" {
		v.AddConfigPath(path)
	}
	v.AddConfigPath(".")
	v.SetConfigName(".env")
	v.SetConfigType("env")

	v.AutomaticEnv()

	// Read config file if exists (ignore error if file not found, fallback to env/defaults)
	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg.App); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(&cfg.Database); err != nil {
		return nil, err
	}

	// Parse allowed origins
	rawOrigins := cfg.App.AllowedOriginsRaw
	if rawOrigins != "" {
		parts := strings.Split(rawOrigins, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cfg.App.AllowedOrigins = append(cfg.App.AllowedOrigins, trimmed)
			}
		}
	}

	if len(cfg.App.AllowedOrigins) == 0 {
		cfg.App.AllowedOrigins = []string{"http://localhost:8081", "http://localhost:5173"}
	}

	return &cfg, nil
}

