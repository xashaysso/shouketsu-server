package cards

import (
	"math/rand/v2"
	"shouketsu/internal/domain"
)

var Catalog = map[string]domain.Card{
	// minions
	"m_peasant": {ID: "m_peasant", Name: "Peasant", Type: "minion", Cost: 1, Attack: 1, Health: 2},
	"m_guard":   {ID: "m_guard", Name: "Guardian", Type: "minion", Cost: 2, Attack: 2, Health: 3},
	"m_berserk": {ID: "m_berserk", Name: "Berserk", Type: "minion", Cost: 3, Attack: 4, Health: 2},
	"m_golem":   {ID: "m_golem", Name: "Golem", Type: "minion", Cost: 4, Attack: 3, Health: 6},

	// spells
	"s_fireball": {ID: "s_fireball", Name: "Fireball", Type: "spell", Cost: 2},
	"s_heal":     {ID: "s_heal", Name: "Heal", Type: "spell", Cost: 1},
}

func GenerateDeck(size int) []domain.Card {
	// filter deck, should also add some field like priority and filter by it later
	var lowCostMinions []domain.Card
	var highCostMinions []domain.Card
	var spells []domain.Card

	for _, card := range Catalog {
		if card.Type == "spell" {
			spells = append(spells, card)
		} else if card.Cost <= 2 {
			lowCostMinions = append(lowCostMinions, card)
		} else {
			highCostMinions = append(highCostMinions, card)
		}
	}

	// shuffle
	deck := make([]domain.Card, 0, size)

	deck = append(deck, pickRandom(lowCostMinions, size/2)...)
	deck = append(deck, pickRandom(highCostMinions, size*3/10)...)
	deck = append(deck, pickRandom(spells, size-len(deck))...) // leftover

	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

func pickRandom(pool []domain.Card, count int) []domain.Card {
	if len(pool) == 0 {
		return nil
	}
	res := make([]domain.Card, 0, count)
	for range count {
		res = append(res, pool[rand.IntN(len(pool))])
	}
	return res
}
