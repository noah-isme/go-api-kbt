package controllers

import (
	"encoding/json"
	"go-api-kbt/config"
	"go-api-kbt/models"
	"net/http"

	"github.com/gorilla/mux"
)

// CreateEvent - Membuat event baru
func CreateEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event
	json.NewDecoder(r.Body).Decode(&event)

	if err := config.DB.Create(&event).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(event)
}

// GetEvents - Mendapatkan daftar semua event
func GetEvents(w http.ResponseWriter, r *http.Request) {
	var events []models.Event
	config.DB.Preload("Users").Find(&events)

	json.NewEncoder(w).Encode(events)
}

// GetEventByID - Mendapatkan event berdasarkan ID
func GetEventByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var event models.Event

	if err := config.DB.Preload("Users").First(&event, params["id"]).Error; err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(event)
}

// JoinEvent - Membiarkan user mengikuti event
func JoinEvent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var user models.User

	// Mendapatkan user berdasarkan ID
	if err := config.DB.First(&user, params["user_id"]).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Memastikan user belum mengikuti event lain
	if user.EventID != nil {
		http.Error(w, "User already joined an event", http.StatusBadRequest)
		return
	}

	// Mendapatkan event berdasarkan ID
	var event models.Event
	if err := config.DB.First(&event, params["event_id"]).Error; err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}
	// Menambahkan user ke event
	user.EventID = &event.ID
	if err := config.DB.Save(&user).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}
