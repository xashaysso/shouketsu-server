package engine

type Card struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Cost   int    `json:"cost"`
	Attack int    `json:"attack,omitempty"`
	Health int    `json:"health,omitempty"`
}

type PlayerState struct {
	ID      string   `json:"id"`
	Health  int      `json:"health"`
	Mana    int      `json:"mana"`
	MaxMana int      `json:"max_mana"`
	Hand    []Card   `json:"hand"`
	Board   [5]*Card `json:"board"`
}

type GameState struct {
	CurrentTurn string                  `json:"current_turn"`
	TurnNumber  int                     `json:"turn_number"`
	Players     map[string]*PlayerState `json:"players"`
}

type Target struct {
	Type string `json:"type"`
	Side string `json:"side"`
	Slot int    `json:"slot,omitempty"`
}

type PlayCardPayload struct {
	CardIndex int     `json:"card_index"`
	SlotIndex int     `json:"slot_index"`
	Target    *Target `json:"target"`
}
