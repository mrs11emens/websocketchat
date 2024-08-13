package server

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func generateClientID() string {
	return uuid.NewString()
}

func broadcastClients() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	clientList := make([]*Client, 0, len(clients))
	for _, client := range clients {
		clientList = append(clientList, client)
	}

	data, err := json.Marshal(map[string]interface{}{
		"mode":    mode,
		"clients": clientList,
	})
	if err != nil {
		fmt.Println("Error marshalling client list:", err)
		return
	}

	for _, client := range clients {
		if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("Error sending client list:", err)
			client.Conn.Close()
			delete(clients, client.ClientID)
		}
	}
}

func sendMessageHistory(conn *websocket.Conn) {
	messageHistoryMu.Lock()
	defer messageHistoryMu.Unlock()

	if len(messageHistory.Messages) == 0 {
		return
	}

	data, err := json.Marshal(messageHistory)
	if err != nil {
		fmt.Println("Error marshalling message history:", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		fmt.Println("Error sending message history:", err)
	}
}