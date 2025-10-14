package activity

import (
	"gorm.io/gorm"
)

// Activity represents an activity in the system.
type Activity struct {
	gorm.Model
	Name        string `gorm:"size:128;not null"`
	Description string `gorm:"type:text"`
}

// TableName overrides the table name for the activity entity.
func (Activity) TableName() string {
	return "activities"
}
