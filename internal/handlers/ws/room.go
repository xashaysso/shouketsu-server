package ws

import (
	"encoding/json"
	"log/slog"
	"shouketsu/internal/domain"
	"shouketsu/internal/engine"
	"time"
)

func RunRoom(r *domain.Room, onEmpty func(roomID string)) {
	slog.Info("room worker started", slog.String("room_id", r.ID))

	gameEngine := engine.NewGameEngine()

	const idleTimeout = 30 * time.Second
	var idleTimer *time.Timer
	var idleChan <-chan time.Time

	for {
		select {
		case client := <-r.Register:

			if idleTimer != nil {
				idleTimer.Stop()
				idleTimer = nil
				idleChan = nil
				slog.Info("room activity resumed, idle timer canceled", slog.String("room_id", r.ID))
			}

			isNameTaken := false
			for c := range r.Clients {
				if c.ID == client.ID {
					isNameTaken = true
					break
				}
			}
			if isNameTaken {
				SendError(client, "this username is already taken in this room")
				continue
			}

			if gameEngine.InProgress {
				_, isParticipant := gameEngine.State.Players[client.ID]
				if !isParticipant {
					SendError(client, "match already in progress in this room")
					continue
				}

				r.Clients[client] = true
				slog.Info("player reconnected", slog.String("id", client.ID))

				view := gameEngine.GetStateForPlayer(client.ID)
				for c := range r.Clients {
					if client.ID != c.ID {
						SendTo(r, c.ID, "PLAYER_RECONNECTED", map[string]string{
							"message": "Opponent reconnected. Game continues...",
						})
					}
				}
				SendTo(r, client.ID, "BOARD_UPDATE", view)
				continue

			}

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

				if gameEngine.InProgress {
					for c := range r.Clients {
						SendTo(r, c.ID, "PLAYER_LEFT", map[string]string{
							"message": "Opponent disconnected. Waiting for reconnect...",
						})
					}
				}
			}
			if len(r.Clients) == 0 {
				slog.Info("room is empty, idle timer started...", slog.String("room_id", r.ID), slog.Duration("timeout", idleTimeout))
				idleTimer = time.NewTimer(idleTimeout)
				idleChan = idleTimer.C
			}
		case <-idleChan:
			if onEmpty != nil {
				slog.Info("room in idle for too long, stopping...", slog.String("room_id", r.ID))
				onEmpty(r.ID)
			}
			return

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
