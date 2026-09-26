package cards

import "shouketsu/internal/domain"

func CastFireball(state *domain.GameState, target *domain.Target, senderID string, opponentID string) error {
	if target == nil {
		return domain.ErrNoTargetProvided
	}

	damage := 3

	if target.Type == "player" {
		targetPlayerID := senderID
		if target.Side == "opponent" {
			targetPlayerID = opponentID
		}
		state.Players[targetPlayerID].Health -= damage
		return nil
	}

	if target.Type == "minion" {
		targetBoard := state.Players[senderID].Board
		if target.Side == "opponent" {
			targetBoard = state.Players[opponentID].Board
		}

		if target.Slot < 0 || target.Slot >= len(targetBoard) || targetBoard[target.Slot] == nil {
			return domain.ErrInvalidCardIndex
		}

		minion := targetBoard[target.Slot]
		minion.Health -= damage

		if minion.Health <= 0 {
			if target.Side == "opponent" {
				state.Players[opponentID].Board[target.Slot] = nil
			} else {
				state.Players[senderID].Board[target.Slot] = nil
			}
		}
		return nil
	}

	return nil
}
