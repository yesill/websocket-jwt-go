package main

import (
	"log"
	"net/http"

	"WebSocket-Go/server"
)

var serv *server.Server

func main() {
	serv = server.NewServer()
	http.HandleFunc("/ws", serv.WSHandler)
	http.HandleFunc("/messages", serv.IncomingMessageHandler)
	log.Println("WebSocket server runs on http://localhost:8000/")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
