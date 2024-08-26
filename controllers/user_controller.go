package controllers

import (
	"encoding/json"
	"go-api-kbt/config"
	"go-api-kbt/models"
	"net/http"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	config.DB.Find(&users)

	json.NewEncoder(w).Encode(users)
}
