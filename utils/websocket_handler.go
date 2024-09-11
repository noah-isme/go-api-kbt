package utils

import (
	"go-api-kbt/config"
	"go-api-kbt/models"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var Clients = make(map[*websocket.Conn]bool)
var Broadcast = make(chan models.Location)

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := config.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	Clients[ws] = true

	for {
		var location models.Location
		err := ws.ReadJSON(&location)
		if err != nil {
			delete(Clients, ws)
			break
		}

		Broadcast <- location
	}
}

func HandleBroadcast() {
	for {
		location := <-Broadcast

		for client := range Clients {
			err := client.WriteJSON(location)
			if err != nil {
				log.Printf("WebSocket error: %v", err)
				client.Close()
				delete(Clients, client)
			}
		}
	}
}

func BroadcastLocation(location models.Location) {
	Broadcast <- location
}
