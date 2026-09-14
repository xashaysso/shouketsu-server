package ws

import (
	"log/slog"
	"shouketsu/internal/entities"
)

func RunRoom(r *entities.Room) {
	slog.Info("room worker started", slog.String("room_id", r.ID))

	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true
			slog.Info("player connected", slog.String("room_id", r.ID), slog.String("player_id", client.ID), slog.Int("total_players", len(r.Clients)))
		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
				slog.Info("player disconnected", slog.String("room_id", r.ID), slog.String("player_id", client.ID))
			}
		case message := <-r.Broadcast:
			for client := range r.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		}
	}
}
