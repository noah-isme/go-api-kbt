package routes

import (
	"go-api-kbt/controllers"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	router := mux.NewRouter()

	// Rute untuk pengguna
	router.HandleFunc("/users", controllers.GetUsers).Methods("GET")
	router.HandleFunc("/users", controllers.CreateUser).Methods("POST")

	return router
}
