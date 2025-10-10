package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config aggregates all configuration groups required by the application.
type Config struct {
	App       AppConfig
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Auth      AuthConfig
	Telemetry TelemetryConfig
}

// AppConfig contains generic application settings.
type AppConfig struct {
	Name        string `envconfig:"NAME" default:"go-api-kbt"`
	Environment string `envconfig:"ENV" default:"development"`
	Version     string `envconfig:"VERSION" default:"0.1.0"`
}

// ServerConfig represents HTTP server specific configuration.
type ServerConfig struct {
	Host                string        `envconfig:"HOST" default:"0.0.0.0"`
	Port                int           `envconfig:"PORT" default:"9090"`
	ReadTimeout         time.Duration `envconfig:"READ_TIMEOUT" default:"15s"`
	WriteTimeout        time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout         time.Duration `envconfig:"IDLE_TIMEOUT" default:"60s"`
	ShutdownGracePeriod time.Duration `envconfig:"SHUTDOWN_GRACE_PERIOD" default:"10s"`
}

// DatabaseConfig defines Postgres connection parameters.
type DatabaseConfig struct {
	Host             string        `envconfig:"HOST" default:"localhost"`
	Port             int           `envconfig:"PORT" default:"5432"`
	User             string        `envconfig:"USER" default:"postgres"`
	Password         string        `envconfig:"PASSWORD" default:"postgres"`
	Name             string        `envconfig:"NAME" default:"go_api_kbt"`
	SSLMode          string        `envconfig:"SSLMODE" default:"disable"`
	MaxOpenConns     int           `envconfig:"MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns     int           `envconfig:"MAX_IDLE_CONNS" default:"10"`
	ConnMaxLifetime  time.Duration `envconfig:"CONN_MAX_LIFETIME" default:"30m"`
	ConnMaxIdleTime  time.Duration `envconfig:"CONN_MAX_IDLE_TIME" default:"5m"`
	MigrationDir     string        `envconfig:"MIGRATION_DIR" default:"migrations"`
	RunMigrations    bool          `envconfig:"RUN_MIGRATIONS" default:"false"`
	SeedAdmin        bool          `envconfig:"SEED_ADMIN" default:"false"`
	StatementTimeout time.Duration `envconfig:"STATEMENT_TIMEOUT" default:"5s"`
}

// RedisConfig represents redis connection configuration.
type RedisConfig struct {
	Addr     string        `envconfig:"ADDR" default:"localhost:6379"`
	Password string        `envconfig:"PASSWORD" default:""`
	DB       int           `envconfig:"DB" default:"0"`
	Enabled  bool          `envconfig:"ENABLED" default:"false"`
	Timeout  time.Duration `envconfig:"TIMEOUT" default:"5s"`
}

// AuthConfig defines JWT and password policy settings.
type AuthConfig struct {
	AccessTokenSecret  string        `envconfig:"ACCESS_TOKEN_SECRET" default:"dev-secret"`
	RefreshTokenSecret string        `envconfig:"REFRESH_TOKEN_SECRET" default:"dev-refresh-secret"`
	AccessTokenTTL     time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL    time.Duration `envconfig:"REFRESH_TOKEN_TTL" default:"720h"`
	HashCost           int           `envconfig:"HASH_COST" default:"12"`
}

// TelemetryConfig encapsulates OpenTelemetry exporter configuration.
type TelemetryConfig struct {
	Endpoint      string `envconfig:"ENDPOINT" default:""`
	Insecure      bool   `envconfig:"INSECURE" default:"true"`
	ServiceName   string `envconfig:"SERVICE_NAME" default:"go-api-kbt"`
	EnableMetrics bool   `envconfig:"ENABLE_METRICS" default:"true"`
	EnableTracing bool   `envconfig:"ENABLE_TRACING" default:"true"`
}

// Load loads configuration values from the environment. When a .env file is
// present it is automatically loaded to ease local development.
func Load(logger *slog.Logger) (*Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			logger.Warn("failed to load .env file", slog.String("error", err.Error()))
		}
	}

	cfg := &Config{}

	loaders := []struct {
		prefix string
		target any
	}{
		{prefix: "APP", target: &cfg.App},
		{prefix: "SERVER", target: &cfg.Server},
		{prefix: "DB", target: &cfg.Database},
		{prefix: "REDIS", target: &cfg.Redis},
		{prefix: "AUTH", target: &cfg.Auth},
		{prefix: "OTEL", target: &cfg.Telemetry},
	}

	for _, l := range loaders {
		if err := envconfig.Process(l.prefix, l.target); err != nil {
			return nil, fmt.Errorf("load config for prefix %s: %w", l.prefix, err)
		}
	}

	cfg.App.Environment = strings.ToLower(cfg.App.Environment)
	return cfg, nil
}
