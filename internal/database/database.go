package database

import (
	"context"
	"fmt"
	"time"

	"go-api-kbt/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
)

// Database abstracts DB operations needed by the application.
type Database interface {
	GormDB() *gorm.DB
	AutoMigrate(dst ...any) error
	Close() error
	DSN() string
}

type postgresDatabase struct {
	db  *gorm.DB
	dsn string
}

// NewPostgres opens a new postgres database connection using the given config.
func NewPostgres(ctx context.Context, cfg *config.Config) (Database, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s", cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode)

	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		return nil, fmt.Errorf("register otelgorm plugin: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &postgresDatabase{db: db, dsn: dsn}, nil
}

func (p *postgresDatabase) GormDB() *gorm.DB {
	return p.db
}

func (p *postgresDatabase) AutoMigrate(dst ...any) error {
	return p.db.AutoMigrate(dst...)
}

func (p *postgresDatabase) Close() error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (p *postgresDatabase) DSN() string {
	return p.dsn
}
