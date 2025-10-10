package event

import (
	"gorm.io/gorm"
)

// Entity models a KBT event stored in the database.
type Entity struct {
	gorm.Model
	Name        string `gorm:"size:128"`
	Description string `gorm:"type:text"`
}

// TableName overrides the table name for the event entity.
func (Entity) TableName() string {
	return "events"
}
