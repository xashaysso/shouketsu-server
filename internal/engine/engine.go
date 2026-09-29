package engine

import (
	"encoding/json"
	"errors"
	"shouketsu/internal/domain"
)

type GameEngine struct {
	InProgress bool
	State      *domain.GameState
}

func NewGameEngine() *GameEngine {
	return &GameEngine{
		InProgress: false,
		State: &domain.GameState{
			Players: make(map[string]*domain.PlayerState),
		},
	}
}

func (e *GameEngine) GetState() *domain.GameState {
	return e.State
}

func (e *GameEngine) DispatchAction(senderID, actionType string, payload json.RawMessage) error {
	if !e.InProgress {
		return domain.ErrGameHasNotStarted
	}

	if senderID != e.State.CurrentTurn {
		return domain.ErrNotYourTurn
	}

	switch actionType {
	case "END_TURN":
		return e.EndTurn(senderID)
	case "PLAY_CARD":
		var p domain.PlayCardPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return errors.New("invalid payload for PLAY_CARD")
		}
		return e.playCard(senderID, p)
	default:
		return errors.New("unknown action type: " + actionType)
	}
}

func (e *GameEngine) CheckWinner() string {
	var dead []string
	var alive []string

	for id, p := range e.State.Players {
		if p.Health <= 0 {
			dead = append(dead, id)
		} else {
			alive = append(alive, id)
		}
	}

	if len(dead) == 2 {
		return "DRAW"
	}
	if len(dead) == 1 && len(alive) == 1 {
		return alive[0]
	}

	return ""
}

func (e *GameEngine) GetStateForPlayer(playerID string) *domain.ClientGameSide {
	me := e.State.Players[playerID]

	var opponent *domain.PlayerState
	for id, p := range e.State.Players {
		if id != playerID {
			opponent = p
			break
		}
	}

	var oppView *domain.OpponentView
	if opponent != nil {
		oppView = &domain.OpponentView{
			ID:        opponent.ID,
			Health:    opponent.Health,
			Mana:      opponent.Mana,
			MaxMana:   opponent.MaxMana,
			HandCount: len(opponent.Hand),
			Board:     opponent.Board,
		}
	}

	return &domain.ClientGameSide{
		CurrentTurn: e.State.CurrentTurn,
		TurnNumber:  e.State.TurnNumber,
		You:         me,
		Opponent:    oppView,
	}
}
