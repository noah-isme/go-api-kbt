package controllers

import (
	"encoding/json"
	"go-api-kbt/config"
	"go-api-kbt/models"
	"go-api-kbt/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func UpdateLocation(w http.ResponseWriter, r *http.Request) {
	var location models.Location
	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Validasi user dan event
	if location.UserID == 0 || location.EventID == 0 {
		http.Error(w, "UserID and EventID are required", http.StatusBadRequest)
		return
	}

	location.Timestamp = time.Now().Unix()

	// Simpan lokasi ke Redis (cache)
	cacheKey := "location_" + strconv.Itoa(int(location.UserID))
	locationData, err := json.Marshal(location)
	if err != nil {
		http.Error(w, "Failed to marshal location data", http.StatusInternalServerError)
		return
	}

	// Periksa apakah Redis client terinisialisasi
	if config.RedisClient == nil {
		log.Println("Redis client is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := config.RedisClient.Set(config.Ctx, cacheKey, locationData, 0).Err(); err != nil {
		http.Error(w, "Failed to set location in Redis", http.StatusInternalServerError)
		return
	}

	// Simpan lokasi ke database
	if err := config.DB.Create(&location).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast ke semua klien via WebSocket
	utils.BroadcastLocation(location)

	json.NewEncoder(w).Encode(location)
}

func GetLiveLocations(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	eventID := params["event_id"]

	// Ambil semua lokasi pengguna dari Redis berdasarkan event_id
	keys := config.RedisClient.Keys(config.Ctx, "location_*").Val()

	var liveLocations []models.Location
	for _, key := range keys {
		locationData, _ := config.RedisClient.Get(config.Ctx, key).Result()
		var location models.Location
		json.Unmarshal([]byte(locationData), &location)
		if strconv.Itoa(int(location.EventID)) == eventID {
			liveLocations = append(liveLocations, location)
		}
	}

	json.NewEncoder(w).Encode(liveLocations)
}
