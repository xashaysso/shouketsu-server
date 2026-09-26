package ws

import (
	"log/slog"
	"shouketsu/internal/domain"

	"github.com/gorilla/websocket"
)

func ReadPump(c *domain.Client) {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Warn("ws read error", slog.String("player", c.ID), slog.Any("err", err))
			}
			break
		}

		c.Room.Actions <- domain.InAction{
			SenderID: c.ID,
			Raw:      message,
		}
	}
}

func WritePump(c *domain.Client) {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			slog.Warn("ws write error", slog.String("player", c.ID), slog.Any("err", err))
			break
		}
	}
}
