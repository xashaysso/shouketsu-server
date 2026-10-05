package ws

import (
	"shouketsu/internal/domain"
	"sync"
)

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*domain.Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*domain.Room),
	}
}

func (m *RoomManager) CreateRoom(roomID string) *domain.Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	if r, ok := m.rooms[roomID]; ok {
		return r
	}

	room := &domain.Room{
		ID:         roomID,
		Clients:    make(map[*domain.Client]bool),
		Actions:    make(chan domain.InAction),
		Register:   make(chan *domain.Client),
		Unregister: make(chan *domain.Client),
	}
	m.rooms[roomID] = room

	go RunRoom(room, func(roomID string) {
		m.DropRoom(roomID)
	})

	return room
}

func (m *RoomManager) DropRoom(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.rooms[roomID]; !ok {
		return ErrNonExistentRoom
	}

	delete(m.rooms, roomID)
	return nil
}
