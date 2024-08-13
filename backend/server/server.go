package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	Conn     *websocket.Conn `json:"-"`
	ClientID string          `json:"client_id"`
	Username string          `json:"username"`
}

type User struct {
	ClientID string `json:"client_id"`
	Username string `json:"username"`
}

type UserList struct {
	Users []User `json:"users"`
}

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

type MessageHistory struct {
	Messages []Message `json:"messages"`
}

var (
	clients         = make(map[string]*Client)
	clientsMu       sync.Mutex
	mode            = "name" // "id" or "name"
	messageHistory  = MessageHistory{}
	messageHistoryMu sync.Mutex
)

const (
	userFile        = "users.json"
	messageFile     = "messages.json"
)

// loadUsers loads the users from the JSON file
func loadUsers() (UserList, error) {
	file, err := ioutil.ReadFile(userFile)
	if err != nil {
		return UserList{}, err
	}

	var users UserList
	err = json.Unmarshal(file, &users)
	if err != nil {
		return UserList{}, err
	}

	return users, nil
}

// saveUsers saves the users to the JSON file
func saveUsers(users UserList) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(userFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// loadMessages loads the message history from the JSON file
func loadMessages() (MessageHistory, error) {
	file, err := ioutil.ReadFile(messageFile)
	if err != nil {
		return MessageHistory{}, err
	}

	var history MessageHistory
	err = json.Unmarshal(file, &history)
	if err != nil {
		return MessageHistory{}, err
	}

	return history, nil
}

// saveMessages saves the message history to the JSON file
func saveMessages(history MessageHistory) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(messageFile, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

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

	// Проверяем, есть ли сообщения в истории
	if len(messageHistory.Messages) == 0 {
		return // Если сообщений нет, не отправляем ничего
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

    // Загрузка пользователей из JSON файла
    users, err := loadUsers()
    if err != nil {
        fmt.Println("Error loading users:", err)
    }

    // Поиск пользователя в списке
    var currentUser *User
    for i, u := range users.Users {
        if u.ClientID == clientID {
            currentUser = &users.Users[i]
            client.Username = u.Username
            break
        }
    }

    // Если пользователь не найден, добавляем его
    if currentUser == nil {
        newUser := User{ClientID: clientID, Username: ""}
        users.Users = append(users.Users, newUser)
        currentUser = &newUser
    }

    // Сохранение обновленного списка пользователей
    err = saveUsers(users)
    if err != nil {
        fmt.Println("Error saving users:", err)
    }

    fmt.Printf("Client %s connected\n", clientID)

    // Отправляем clientID обратно клиенту
    if err := conn.WriteMessage(websocket.TextMessage, []byte(clientID)); err != nil {
        fmt.Println("Error sending client ID:", err)
    }

    // Отправляем историю сообщений клиенту
    sendMessageHistory(conn)

    // Broadcast client list after connection
    broadcastClients()

    // Слушаем сообщения от клиента
    for {
        messageType, msg, err := conn.ReadMessage()
        if err != nil {
            fmt.Println("Error while reading message:", err)
            break
        }

        // Обработка установки никнейма
        if string(msg[:5]) == "/nick" {
            newName := string(msg[6:])
            client.Username = newName
            currentUser.Username = newName
            fmt.Printf("Client %s set name to %s\n", clientID, client.Username)

            // Сохранение обновленного списка пользователей
            err = saveUsers(users)
            if err != nil {
                fmt.Println("Error saving users:", err)
            }

            // Broadcast updated client list
            broadcastClients()
            continue
        }

        // Переключение режима отображения
        if string(msg) == "/toggle" {
            if mode == "id" {
                mode = "name"
            } else {
                mode = "id"
            }
            msg = []byte("Mode changed to " + mode)
        }

        // Форматирование сообщений
        var displayID string
        if mode == "name" && client.Username != "" {
            displayID = client.Username
        } else {
            displayID = client.ClientID
        }

        // Добавление сообщения в историю
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