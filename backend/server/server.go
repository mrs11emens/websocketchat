package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = generateClientID()
	}

	clientsMu.Lock()
	client := &Client{
		Conn:     conn,
		ClientID: clientID,
		Username: "",
	}
	clients[clientID] = client
	clientsMu.Unlock()

	users, err := loadUsers()
	if err != nil {
		fmt.Println("Error loading users:", err)
	}

	var currentUser *User
	for i, u := range users.Users {
		if u.ClientID == clientID {
			currentUser = &users.Users[i]
			client.Username = u.Username
			break
		}
	}

	if currentUser == nil {
		newUser := User{ClientID: clientID, Username: ""}
		users.Users = append(users.Users, newUser)
		currentUser = &newUser
	}

	err = saveUsers(users)
	if err != nil {
		fmt.Println("Error saving users:", err)
	}

	fmt.Printf("Client %s connected\n", clientID)

	if err := conn.WriteMessage(websocket.TextMessage, []byte(clientID)); err != nil {
		fmt.Println("Error sending client ID:", err)
	}

	sendMessageHistory(conn)
	broadcastClients()

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		if string(msg[:5]) == "/nick" {
			newName := string(msg[6:])
			client.Username = newName
			currentUser.Username = newName
			fmt.Printf("Client %s set name to %s\n", clientID, client.Username)

			err = saveUsers(users)
			if err != nil {
				fmt.Println("Error saving users:", err)
			}

			broadcastClients()
			continue
		}

		if string(msg) == "/toggle" {
			if mode == "id" {
				mode = "name"
			} else {
				mode = "id"
			}
			msg = []byte("Mode changed to " + mode)
		}

		var displayID string
		if mode == "name" && client.Username != "" {
			displayID = client.Username
		} else {
			displayID = client.ClientID
		}

		messageHistoryMu.Lock()
		messageHistory.Messages = append(messageHistory.Messages, Message{
			Sender:  displayID,
			Content: string(msg),
		})
		saveMessages(messageHistory)
		messageHistoryMu.Unlock()

		fmt.Printf("Received from %s: %s\n", clientID, msg)

		clientsMu.Lock()
		for _, c := range clients {
			if err := c.Conn.WriteMessage(messageType, []byte(fmt.Sprintf("[%s] %s", displayID, msg))); err != nil {
				fmt.Println("Error while writing message:", err)
				c.Conn.Close()
				delete(clients, c.ClientID)
			}
		}
		clientsMu.Unlock()
	}
}