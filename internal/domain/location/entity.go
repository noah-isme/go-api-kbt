package location

import "gorm.io/gorm"

// Entity represents a tracked location for a user during an event.
type Entity struct {
	gorm.Model
	UserID    uint
	EventID   uint
	Latitude  float64
	Longitude float64
	Timestamp int64
}

// TableName overrides the locations table name.
func (Entity) TableName() string {
	return "locations"
}
