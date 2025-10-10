package user

import (
	"time"

	"gorm.io/gorm"
)

// Role enumerates supported user roles.
type Role string

const (
	RoleOwner   Role = "owner"
	RoleAdmin   Role = "admin"
	RoleCashier Role = "cashier"
)

// Entity represents the persistent model for a system user.
type Entity struct {
	gorm.Model
	Username    string `gorm:"uniqueIndex;size:64"`
	Email       string `gorm:"uniqueIndex;size:255"`
	Password    string `gorm:"size:255"`
	Role        Role   `gorm:"size:32"`
	EventID     *uint
	LastLoginAt *time.Time
}

// TableName overrides the default table name for the Entity.
func (Entity) TableName() string {
	return "users"
}
