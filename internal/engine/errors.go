package engine

import (
	"errors"
)

var (
	ErrInvalidCardIndex    = errors.New("invalid card index")
	ErrNotYourTurn         = errors.New("not your turn")
	ErrGameHasNotStarted   = errors.New("game has not started yet")
	ErrNotEnoughMana       = errors.New("not enough mana")
	ErrInvalidCardType     = errors.New("invalid card type")
	ErrInvalidSlotIndex    = errors.New("invalid slot index")
	ErrSlotAlreadyOccupied = errors.New("slot already occupied")
	ErrUnknownSpell        = errors.New("unknown spell effect")
	ErrNoTargetProvided    = errors.New("this card requires target")
)
