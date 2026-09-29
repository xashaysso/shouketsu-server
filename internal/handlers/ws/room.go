package ws

import (
	"encoding/json"
	"log/slog"
	"shouketsu/internal/domain"
	"shouketsu/internal/engine"
)

func RunRoom(r *domain.Room) {
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

				for client := range r.Clients {
					SendTo(r, client.ID, "GAME_STARTED", gameEngine.GetStateForPlayer(client.ID))
				}
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

func Broadcast(r *domain.Room, msgType string, payload any) {
	bytes, err := json.Marshal(domain.OutMessage{
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

func BroadcastGameState(r *domain.Room, engine *engine.GameEngine) {
	for client := range r.Clients {
		view := engine.GetStateForPlayer(client.ID)
		SendTo(r, client.ID, "BOARD_UPDATE", view)
	}
}

func SendTo(r *domain.Room, targetID, msgType string, payload any) {
	bytes, err := json.Marshal(domain.OutMessage{
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

func HandleAction(r *domain.Room, engine *engine.GameEngine, action domain.InAction) {
	var env domain.ActionEnvelope
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

	BroadcastGameState(r, engine)

	if winner := engine.CheckWinner(); winner != "" {
		engine.InProgress = false

		Broadcast(r, "GAME_OVER", map[string]any{
			"winner_id": winner,
			"is_draw":   winner == "DRAW",
		})

		slog.Info("game finished", slog.String("room_id", r.ID), slog.String("winner", winner))
	}
}
