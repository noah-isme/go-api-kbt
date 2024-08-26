package main

import (
	"go-api-kbt/config"
	"go-api-kbt/routes"
	"log"
	"net/http"
)

func main() {
	// Inisialisasi database
	config.InitDB()

	// Inisialisasi router
	router := routes.InitRoutes()

	log.Println("Starting server on :9090")
	log.Fatal(http.ListenAndServe(":9090", router))
}
