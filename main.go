package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type DataProps struct {
	CaseID    string `json:"case_id"`
	PatientId string `json:"patient_id omitempty"`
}

type User struct {
	ID   string
	Conn *websocket.Conn
}

type UserMessage struct {
	Action string    `json:"action"`
	Data   DataProps `json:"data"`
}

var actions = map[string]string{
	"OPEN_CASE":    "CASE_OPENED",
	"OPEN_PATIENT": "PATIENT_OPENED",
}

var (
	users = make(map[string]*User)
	mu    sync.Mutex
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		// Set example variable
		c.Set("example", "12345")
		// token := c.Request.Header["Token"]
		// log.Println("....token...", token[0])

		// before request

		c.Next()

		// after request
		latency := time.Since(t)
		log.Print(latency)

		// access the status we are sending
		status := c.Writer.Status()
		log.Println(status)
	}
}

func main() {

	// Create a new Gin router
	r := gin.Default()
	r.Use(Logger())

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

		// defer conn.Close()
	}
	err := conn.WriteMessage(websocket.TextMessage, []byte(actions[action]))
	if err != nil {
		log.Printf("Error %s when Repling message to client", err)
		return err
	}
	return nil
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	// In a real application, you would authenticate the user and get their userID
	userID := c.Query("user_id")
	if userID == "" {
		log.Println("User ID is missing")
		return
	}

	mu.Lock()
	users[userID] = &User{ID: userID, Conn: conn}
	mu.Unlock()

	log.Printf("User %s connected", userID)

	for {
		var msg map[string]string
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error: %v", err)
			break
		}

		targetID := msg["targetID"]
		message := msg["message"]

		mu.Lock()
		targetUser, ok := users[targetID]
		mu.Unlock()

		if ok {
			err := targetUser.Conn.WriteJSON(map[string]string{"message": message})
			if err != nil {
				log.Printf("Error: %v", err)
			}
		} else {
			log.Printf("User %s not connected", targetID)
		}
	}

	mu.Lock()
	delete(users, userID)
	mu.Unlock()
	log.Printf("User %s disconnected", userID)
}

// func wsHandler(c *gin.Context) {
// 	userId := c.Query("user_id")

// 	if userId == "" {
// 		log.Println("User ID is missing")
// 		return
// 	}

// 	// Upgrade the HTTP connection to a WebSocket connection
// 	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
// 	if err != nil {
// 		log.Println("Failed to set websocket upgrade:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	mu.Lock()
// 	users[userId] = &Client{ID: userId, Conn: conn}
// 	mu.Unlock()

// 	log.Printf("User %s connected", userId)

// 	// Handle WebSocket messages
// 	for {
// 		var msg map[string]string
// 		err := conn.ReadJSON(&msg)
// 		if err != nil {
// 			log.Printf("Error: %v", err)
// 			break
// 		}
// 		// starte read message
// 		// here messageType is an int with value websocket.BinaryMessage or websocket.TextMessage
// 		// message user message
// 		messageType, message, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println("Read error:", err)
// 			return
// 		}
// 		if messageType == websocket.BinaryMessage {
// 			log.Println("we unable to read binary message")
// 			return
// 		}

// 		if messageType == websocket.TextMessage {
// 			userMessage := UserMessage{}
// 			log.Printf("isJSON...%v\n", isJSON(string(message)))
// 			if isJSON(string(message)) {
// 				err := json.Unmarshal(message, &userMessage)
// 				if err != nil {
// 					log.Println("Unmarshal error:", err)
// 					return
// 				}
// 				action := strings.ToUpper(strings.Trim(userMessage.Action, "\n"))
// 				err = sendMessage(conn, action)
// 				log.Println("error:", err)

// 				return
// 			} else {
// 				action := strings.ToUpper(strings.Trim(string(message), "\n"))
// 				err := sendMessage(conn, action)
// 				log.Println("error:", err)

// 			}

// 		}
// 	}
// }
