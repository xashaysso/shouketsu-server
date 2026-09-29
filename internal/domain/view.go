package domain

type OpponentView struct {
	ID        string               `json:"id"`
	Health    int                  `json:"health"`
	Mana      int                  `json:"mana"`
	MaxMana   int                  `json:"max_mana"`
	HandCount int                  `json:"hand_count"`
	Board     [MaxBoardSlots]*Card `json:"board"`
}

type ClientGameSide struct {
	CurrentTurn string        `json:"current_turn"`
	TurnNumber  int           `json:"turn_number"`
	You         *PlayerState  `json:"you"`
	Opponent    *OpponentView `json:"opponent"`
}
