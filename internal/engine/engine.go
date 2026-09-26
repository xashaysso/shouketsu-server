package engine

import (
	"encoding/json"
	"errors"
	"shouketsu/internal/combat"
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

func (e *GameEngine) StartGame(p1ID, p2ID string) {
	e.InProgress = true
	e.State.CurrentTurn = p1ID
	e.State.TurnNumber = 1

	starterDeck := []domain.Card{
		{ID: "c1", Name: "card1", Type: "minion", Cost: 1, Attack: 1, Health: 2},
		{ID: "spell_fireball", Name: "fireball", Type: "spell", Cost: 1},
		{ID: "c3", Name: "card3", Type: "minion", Cost: 1, Attack: 2, Health: 1},
	}

	e.State.Players[p1ID] = &domain.PlayerState{
		ID:      p1ID,
		Health:  30,
		Mana:    1,
		MaxMana: 1,
		Hand:    append([]domain.Card{}, starterDeck...),
	}
	e.State.Players[p2ID] = &domain.PlayerState{
		ID:      p2ID,
		Health:  30,
		Mana:    1,
		MaxMana: 1,
		Hand:    append([]domain.Card{}, starterDeck...),
	}
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

func (e *GameEngine) EndTurn(senderID string) error {
	combat.ResolveCombatPhase(e.GetState(), senderID)

	var nextTurnPlayer string

	for id := range e.State.Players {
		if id != senderID {
			nextTurnPlayer = id
			break
		}
	}

	e.State.CurrentTurn = nextTurnPlayer
	e.State.TurnNumber++

	nextPlayer := e.State.Players[nextTurnPlayer]
	if nextPlayer.MaxMana < 10 {
		nextPlayer.MaxMana++
	}
	nextPlayer.Mana = nextPlayer.MaxMana

	return nil
}

func (e *GameEngine) playCard(senderID string, p domain.PlayCardPayload) error {
	player := e.State.Players[senderID]

	if p.CardIndex < 0 || p.CardIndex >= len(player.Hand) {
		return domain.ErrInvalidCardIndex
	}

	card := player.Hand[p.CardIndex]
	if player.Mana < card.Cost {
		return domain.ErrNotEnoughMana
	}

	var err error
	switch card.Type {
	case "minion":
		err = e.PlayMinion(player, card, p.SlotIndex)
	case "spell":
		err = e.PlaySpell(senderID, card, p.Target)
	default:
		return domain.ErrInvalidCardType
	}

	if err != nil {
		return err
	}

	player.Mana -= card.Cost
	player.Hand = append(player.Hand[:p.CardIndex], player.Hand[p.CardIndex+1:]...)

	return nil
}
