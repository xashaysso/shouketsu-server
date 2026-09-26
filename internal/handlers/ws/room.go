package ws

import (
	"encoding/json"
	"log/slog"
	"shouketsu/internal/engine"
	"shouketsu/internal/entities"
)

func RunRoom(r *entities.Room) {
	slog.Info("room worker started", slog.String("room_id", r.ID))

	gameEngine := engine.NewGameEngine()

	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true
			slog.Info("player connected", slog.String("room_id", r.ID), slog.String("player_id", client.ID), slog.Int("total_players", len(r.Clients)))

			if len(r.Clients) == 2 && !gameEngine.InProgress {
				players := make([]string, 0, 2)
				for c := range r.Clients {
					players = append(players, c.ID)
				}
				gameEngine.StartGame(players[0], players[1])
				Broadcast(r, "GAME_STARTED", gameEngine.GetState())
			}
		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
				slog.Info("player disconnected", slog.String("room_id", r.ID), slog.String("player_id", client.ID))
			}
		case action := <-r.Actions:
			HandleAction(r, gameEngine, action)
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

func HandleAction(r *entities.Room, engine *engine.GameEngine, action entities.InAction) {
	var env entities.ActionEnvelope
	if err := json.Unmarshal(action.Raw, &env); err != nil {
		SendTo(r, action.SenderID, "ERROR", map[string]string{
			"message": "invalid JSON format",
		})
		return
	}

	err := engine.DispatchAction(action.SenderID, env.Type, env.Payload)
	if err != nil {
		SendTo(r, action.SenderID, "ERROR", map[string]string{
			"message": err.Error(),
		})
		return
	}

	Broadcast(r, "BOARD_UPDATE", engine.GetState())
}
