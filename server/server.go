package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var allowed_headers = []string{"Token", "content-type"}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	connections map[*websocket.Conn]string
}

func NewServer() *Server {
	/* Creates a new Server with empty conenctions map. */

	return &Server{
		connections: make(map[*websocket.Conn]string),
	}
}

func (server *Server) WSHandler(w http.ResponseWriter, r *http.Request) {
	/* Handles WebSocket connections, creates JWT Tokens per connection and sends it as response. */

	w = SetCORS(w)
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("error in WSHandler: ", err)
		return
	}
	/* defer connection.Close() */

	fmt.Println("new connection:", connection.RemoteAddr())

	initialPlayerID := connection.RemoteAddr().String()
	//add incoming connection to the connections
	server.connections[connection] = initialPlayerID

	//JWT token creation start
	token, err := GenerateJWT(initialPlayerID)
	if err != nil {
		return
	}
	//JWT token creation end
	//initial message
	server.writeMessage(-2, initialPlayerID, token, connection)
	message := fmt.Sprintf("Your Initial Player ID: %s\nIf you want to change it type '/cid *new player id*'", initialPlayerID)
	server.writeMessage(-1, initialPlayerID, message, connection)
}

func (server *Server) IncomingMessageHandler(w http.ResponseWriter, r *http.Request) {
	/* Handles incoming messages. */

	w = SetCORS(w)
	if r.Method == "POST" {
		if r.Header[allowed_headers[0]] != nil {
			// if JWT token is valid
			token := r.Header[allowed_headers[0]][0]
			if ValidateJWT(token) {
				senderPlayerID := GetPlayerIDFromJWTToken(token)
				receivedData := decodeReceivedData(w, r)
				if !server.forwardMessage(senderPlayerID, receivedData) {
					fmt.Fprint(w, "Error: Failed to forward message!")
				}
			} else {
				println("Not Authorized: Token is not valid!")
			}
		} else {
			println("Error: Header 'Token' not found in request!")
		}
	} else {
		fmt.Fprint(w, "Error: POST method only!")
	}
}

func decodeReceivedData(w http.ResponseWriter, r *http.Request) CommunicationData {
	/* Decodes received JSON data from client to a 'CommunicationData' and returns it. */

	decoder := json.NewDecoder(r.Body)
	var commData CommunicationData
	err := decoder.Decode(&commData)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
	}
	return commData
}

func (server *Server) changePlayerID(senderPlayerID string, commData CommunicationData) {
	/* Changes PlayerId, if the intended player id exists in the connections no change will be done. */

	intendedPlayerID := commData.Message
	connection := server.findClientByPlayerID(intendedPlayerID)
	if connection == nil {
		connection = server.findClientByPlayerID(senderPlayerID)
		server.connections[connection] = intendedPlayerID
		//new JWT Token
		token, err := GenerateJWT(server.connections[connection])
		if err != nil {
			return
		}
		newPlayerID := server.connections[connection]
		server.writeMessage(-2, newPlayerID, token, connection)

		message := fmt.Sprintf("Your new PlayerID: %s", newPlayerID)
		server.writeMessage(-1, newPlayerID, message, connection)
		//optional: new player id can be broadcasted so the other players know who changed player id and what is it now
	} else {
		server.writeMessage(-1, senderPlayerID, "PlayerID already exists!", server.findClientByPlayerID(senderPlayerID))
	}
}

func (server *Server) forwardMessage(senderPlayerID string, commData CommunicationData) bool {
	/* Forwards the message inside the retrieved data to the intended receiver(client). */

	if commData.Type == 0 { //broadcast
		for c := range server.connections {
			go func(c *websocket.Conn) {
				server.writeMessage(commData.Type, senderPlayerID, commData.Message, c)
			}(c)
		}
		return true
	} else if commData.Type == 1 { //whisper
		connectionSender := server.findClientByPlayerID(senderPlayerID)
		connectionReceiver := server.findClientByPlayerID(commData.PlayerID)
		if connectionSender != nil {
			message := fmt.Sprintf("Whisper to %s: %s", commData.PlayerID, commData.Message)
			server.writeMessage(commData.Type, senderPlayerID, message, connectionSender)
		} else {
			return false
		}
		if connectionReceiver != nil {
			message := fmt.Sprintf("Whisper from %s: %s", senderPlayerID, commData.Message)
			server.writeMessage(commData.Type, senderPlayerID, message, connectionReceiver)
		} else {
			return false
		}
		return true

	} else if commData.Type == -1 {
		server.changePlayerID(senderPlayerID, commData)
		return true
	} else { // type should be 0 or 1
		return false
	}
}

func (server *Server) writeMessage(messageType int, senderPlayerID string, message string, connection *websocket.Conn) {
	/* Writes a message in JSON format to a connection. */

	/*
		types:
		-2 -> message is a JWT Token
		-1 -> it is a system message
		0 -> broadcast message
		1 -> client to client whisper message
	*/
	sentJsonData, err := json.Marshal(
		CommunicationData{
			Type:     messageType,
			PlayerID: senderPlayerID,
			Message:  message,
		},
	)
	if err != nil {
		log.Println("Failed to marshal JSON:", err)
		return
	}

	err = connection.WriteMessage(websocket.TextMessage, sentJsonData)
	if err != nil {
		if strings.Contains(err.Error(), "wsasend: An established connection was aborted by the software in your host machine") {
			server.closeAndDeleteConnection(connection)
		} else {
			fmt.Println("write error: ", err)
		}
	}
}

func (server *Server) findClientByPlayerID(playerID string) *websocket.Conn {
	/* Finds specific client in connections map by it's remote address (websocket.Conn.RemoteAddr()). */

	for connection, pid := range server.connections {
		if pid == playerID {
			return connection
		}
	}
	return nil
}

func (server *Server) closeAndDeleteConnection(connection *websocket.Conn) {
	/* Closes a connection and deletes it from the connections map. */

	connection.Close()
	delete(server.connections, connection)
}

func SetCORS(w http.ResponseWriter) http.ResponseWriter {
	/* Sets CORS(Cross-Origin Resource Sharing) headers for a smooth experience. */

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowed_headers, ","))
	return w
}
