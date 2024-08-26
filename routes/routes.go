package routes

import (
	"go-api-kbt/controllers"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	// Rute untuk pengguna
	router.HandleFunc("/users", controllers.GetUsers).Methods("GET")
	router.HandleFunc("/users/{id}", controllers.GetUserByID).Methods("GET")
	router.HandleFunc("/users", controllers.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", controllers.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", controllers.DeleteUser).Methods("DELETE")

	// Rute untuk event
	router.HandleFunc("/events", controllers.GetEvents).Methods("GET")
	router.HandleFunc("/events/{id}", controllers.GetEventByID).Methods("GET")
	router.HandleFunc("/events", controllers.CreateEvent).Methods("POST")

	// Rute untuk mengikuti event
	router.HandleFunc("/events/{event_id}/join/{user_id}", controllers.JoinEvent).Methods("POST")

	return router
}
