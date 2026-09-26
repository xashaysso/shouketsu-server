package engine

func (e *GameEngine) playSpell(senderID string, card Card, target *Target) error {
	var opponentID string
	for id := range e.State.Players {
		if id != senderID {
			opponentID = id
			break
		}
	}

	switch card.ID {
	case "spell_fireball":
		return e.playFireball(target, senderID, opponentID)
	default:
		return ErrUnknownSpell
	}
}

func (e *GameEngine) playFireball(target *Target, senderID string, opponentID string) error {
	if target == nil {
		return ErrNoTargetProvided
	}

	damage := 3

	if target.Type == "player" {
		targetPlayerID := senderID
		if target.Side == "opponent" {
			targetPlayerID = opponentID
		}
		e.State.Players[targetPlayerID].Health -= damage
		return nil
	}

	if target.Type == "minion" {
		targetBoard := e.State.Players[senderID].Board
		if target.Side == "opponent" {
			targetBoard = e.State.Players[opponentID].Board
		}

		if target.Slot < 0 || target.Slot >= len(targetBoard) || targetBoard[target.Slot] == nil {
			return ErrInvalidCardIndex
		}

		minion := targetBoard[target.Slot]
		minion.Health -= damage

		if minion.Health <= 0 {
			if target.Side == "opponent" {
				e.State.Players[opponentID].Board[target.Slot] = nil
			} else {
				e.State.Players[senderID].Board[target.Slot] = nil
			}
		}
		return nil
	}

	return nil
}
