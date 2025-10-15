package location

import "gorm.io/gorm"

// Entity represents a stored location such as a meetup or checkpoint.
type Entity struct {
	gorm.Model
	Name    string
	Address string
}

// TableName overrides the locations table name.
func (Entity) TableName() string {
	return "locations"
}
