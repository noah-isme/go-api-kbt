package models

import "gorm.io/gorm"

type Event struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Users       []User `json:"users" gorm:"foreignKey:EventID"`
}

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"password"`
	EventID  *uint  `json:"event_id"` // Foreign key ke Event
}
