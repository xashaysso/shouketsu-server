package combat

import "shouketsu/internal/domain"

func ResolveCombatPhase(state *domain.GameState, senderID string) error {
	var opponentID string
	for id := range state.Players {
		if id != senderID {
			opponentID = id
			break
		}
	}

	for id, card := range state.Players[senderID].Board {
		if card == nil {
			continue
		}
		err := AttackSlot(state, id, senderID, opponentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func AttackSlot(state *domain.GameState, slot int, senderID, opponentID string) error {
	if slot < 0 || slot >= len(state.Players[senderID].Board) {
		return domain.ErrInvalidSlotIndex
	}

	senderCard := state.Players[senderID].Board[slot]
	if senderCard == nil {
		return nil
	}

	opponentCard := state.Players[opponentID].Board[slot]
	if opponentCard == nil {
		state.Players[opponentID].Health -= senderCard.Attack
	} else {
		opponentCard.Health -= senderCard.Attack
		senderCard.Health -= opponentCard.Attack
		if opponentCard.Health <= 0 {
			state.Players[opponentID].Board[slot] = nil
		}
		if senderCard.Health <= 0 {
			state.Players[senderID].Board[slot] = nil
		}
	}

	return nil
}
