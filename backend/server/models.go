package server

import (
	"sync"

	"github.com/gorilla/websocket"
)

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