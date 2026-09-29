package engine

import (
	"encoding/json"
	"errors"
	"shouketsu/internal/cards"
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

	p1Deck := cards.GenerateDeck(domain.MaxDeckSize)
	p2Deck := cards.GenerateDeck(domain.MaxDeckSize)

	e.State.Players[p1ID] = &domain.PlayerState{
		ID:      p1ID,
		Health:  30,
		Mana:    1,
		MaxMana: 1,
		Deck:    p1Deck,
		Hand:    make([]domain.Card, 0, domain.MaxHandSize),
	}
	e.State.Players[p2ID] = &domain.PlayerState{
		ID:      p2ID,
		Health:  30,
		Mana:    1,
		MaxMana: 1,
		Deck:    p2Deck,
		Hand:    make([]domain.Card, 0, domain.MaxHandSize),
	}

	for range domain.StartHandSize {
		e.DrawCard(p1ID)
		e.DrawCard(p2ID)
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

	if e.CheckWinner() != "" {
		return nil
	}

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
	if nextPlayer.MaxMana < domain.MaxManaCap {
		nextPlayer.MaxMana++
	}
	nextPlayer.Mana = nextPlayer.MaxMana

	e.DrawCard(nextTurnPlayer)

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

func (e *GameEngine) DrawCard(playerID string) error {
	player := e.State.Players[playerID]

	if len(player.Deck) == 0 {
		player.Health -= 1
		return nil
	}

	card := player.Deck[0]
	player.Deck = player.Deck[1:]

	if len(player.Hand) >= domain.MaxHandSize {
		return nil
	}

	player.Hand = append(player.Hand, card)
	return nil
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
