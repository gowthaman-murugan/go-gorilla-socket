package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type DataProps struct {
	CaseID    string `json:"case_id"`
	PatientId string `json:"patient_id omitempty"`
}

type UserMessage struct {
	Action string    `json:"action"`
	Data   DataProps `json:"data"`
}

var actions = map[string]string{
	"OPEN_CASE":    "CASE_OPENED",
	"OPEN_PATIENT": "PATIENT_OPENED",
}

func main() {

	// Create a new Gin router
	r := gin.Default()

	r.GET("/ws", wsHandler)

	// Start Gin server
	if err := r.Run("localhost:8080"); err != nil {
		log.Fatal("Server Run Failed:", err)
	}
	log.Print("Starting server...")

}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

func sendMessage(conn *websocket.Conn, action string) error {
	if actions[action] == "" {
		err := conn.WriteMessage(websocket.TextMessage, []byte("ACTION_NOT_ALLOWED"))
		if err != nil {
			log.Printf("Error %s when sending message to client", err)
			return err
		}

		defer conn.Close()
	}
	err := conn.WriteMessage(websocket.TextMessage, []byte(actions[action]))
	if err != nil {
		log.Printf("Error %s when Repling message to client", err)
		return err
	}
	return nil
}

func wsHandler(c *gin.Context) {

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close()

	// Handle WebSocket messages
	for {
		// starte read message
		// here messageType is an int with value websocket.BinaryMessage or websocket.TextMessage
		// message user message
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}
		if messageType == websocket.BinaryMessage {
			log.Println("we unable to read binary message")
			return
		}

		if messageType == websocket.TextMessage {
			userMessage := UserMessage{}
			log.Printf("isJSON...%v\n", isJSON(string(message)))
			if isJSON(string(message)) {
				err := json.Unmarshal(message, &userMessage)
				if err != nil {
					log.Println("Unmarshal error:", err)
					return
				}
				action := strings.ToUpper(strings.Trim(userMessage.Action, "\n"))
				err = sendMessage(conn, action)
				log.Println("error:", err)

				return
			} else {
				action := strings.ToUpper(strings.Trim(string(message), "\n"))
				err := sendMessage(conn, action)
				log.Println("error:", err)

			}

		}
	}
}
