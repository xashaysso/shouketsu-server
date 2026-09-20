package entities

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	ID   string
	Room *Room
	Conn *websocket.Conn
	Send chan []byte
}

type Room struct {
	ID         string
	Clients    map[*Client]bool
	Actions    chan InAction
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

func NewRoom(id string) *Room {
	return &Room{
		ID:         id,
		Clients:    make(map[*Client]bool),
		Actions:    make(chan InAction),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

type InAction struct {
	SenderID string
	Raw      []byte
}

type ActionEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type OutMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
