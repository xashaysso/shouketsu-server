package ws

import (
	"encoding/json"
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
		case action := <-r.Actions:
			HandleAction(r, action)
		}
	}
}

func Broadcast(r *entities.Room, msgType string, payload any) {
	bytes, err := json.Marshal(entities.OutMessage{
		Type:    msgType,
		Payload: payload,
	})

	if err != nil {
		slog.Error("failed to marshal broadcast", slog.Any("err", err))
		return
	}

	for client := range r.Clients {
		select {
		case client.Send <- bytes:
		default:
			close(client.Send)
			delete(r.Clients, client)
		}
	}
}

func SendTo(r *entities.Room, targetID, msgType string, payload any) {
	bytes, err := json.Marshal(entities.OutMessage{
		Type:    msgType,
		Payload: payload,
	})
	if err != nil {
		slog.Error("failed to marshal message", slog.Any("err", err))
		return
	}

	for client := range r.Clients {
		if client.ID == targetID {
			select {
			case client.Send <- bytes:
			default:
				close(client.Send)
				delete(r.Clients, client)
			}
			return
		}
	}
}

func HandleAction(r *entities.Room, action entities.InAction) {
	var env entities.ActionEnvelope
	if err := json.Unmarshal(action.Raw, &env); err != nil {
		SendTo(r, action.SenderID, "ERROR", map[string]string{
			"message": "invalid JSON format",
		})
		return
	}

	switch env.Type {
	case "PING":
		SendTo(r, action.SenderID, "PONG", map[string]string{
			"status": "alive",
		})
	default:
		SendTo(r, action.SenderID, "ERROR", map[string]string{
			"message": "unknown action: " + env.Type,
		})
	}
}
