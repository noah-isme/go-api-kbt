package database

import (
	"context"
	"fmt"

	"go-api-kbt/internal/config"
	domainUser "go-api-kbt/internal/domain/user"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin creates an admin user if requested.
func SeedAdmin(ctx context.Context, db Database, cfg *config.Config) error {
	if !cfg.Database.SeedAdmin {
		return nil
	}

	gormDB := db.GormDB()
	var count int64
	if err := gormDB.WithContext(ctx).Model(&domainUser.Entity{}).Where("role = ?", domainUser.RoleOwner).Count(&count).Error; err != nil {
		return fmt.Errorf("count admin users: %w", err)
	}

	if count > 0 {
		return nil
	}

	password := "ChangeMe123!"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), cfg.Auth.HashCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	admin := &domainUser.Entity{
		Username: "admin",
		Email:    "admin@example.com",
		Password: string(hashed),
		Role:     domainUser.RoleOwner,
	}

	if err := gormDB.WithContext(ctx).Create(admin).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	return nil
}
