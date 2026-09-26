package engine

import (
	"shouketsu/internal/cards"
	"shouketsu/internal/domain"
)

func (e *GameEngine) PlaySpell(senderID string, card domain.Card, target *domain.Target) error {
	var opponentID string
	for id := range e.State.Players {
		if id != senderID {
			opponentID = id
			break
		}
	}

	switch card.ID {
	case "s_fireball":
		return cards.CastFireball(e.State, target, senderID, opponentID)
	default:
		return domain.ErrUnknownSpell
	}
}

func (e *GameEngine) PlayMinion(player *domain.PlayerState, card domain.Card, slotIndex int) error {
	if slotIndex < 0 || slotIndex >= len(player.Board) {
		return domain.ErrInvalidSlotIndex
	}
	if player.Board[slotIndex] != nil {
		return domain.ErrSlotAlreadyOccupied
	}

	cardCopy := card
	player.Board[slotIndex] = &cardCopy

	return nil
}
