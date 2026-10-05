package ws

import (
	"log/slog"
	"net/http"
	"shouketsu/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WShandler struct {
	roomManager *RoomManager
}

func NewWShandler(m *RoomManager) *WShandler {
	return &WShandler{
		roomManager: m,
	}
}

func (h *WShandler) ServeWS(c *gin.Context) {
	playerID := c.Query("player")
	if playerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query param 'player' required",
		})
		return
	}
	roomID := c.Query("room")
	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query param 'room' required", // will be changed later on to random matchmaking UUID
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("failed to upgrade to websocket", slog.Any("err", err))
		return
	}

	room := h.roomManager.CreateRoom(roomID)

	client := &domain.Client{
		ID:   playerID,
		Room: room,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Room.Register <- client

	go WritePump(client)
	go ReadPump(client)
}
