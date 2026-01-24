package utils

import (
	"strconv"
	"strings"

	"github.com/vgaro/yotocli/pkg/yoto"
)

// FindCard searches for a card by index (1-based), ID, or Title substring.
func FindCard(cards []yoto.Card, query string) *yoto.Card {
	// Try Index
	if idx, err := strconv.Atoi(query); err == nil {
		if idx > 0 && idx <= len(cards) {
			return &cards[idx-1]
		}
	}

	// Try ID match
	for _, card := range cards {
		if card.CardID == query {
			return &card
		}
	}

	// Try Title match (case-insensitive substring)
	queryLower := strings.ToLower(query)
	for _, card := range cards {
		if strings.Contains(strings.ToLower(card.Title), queryLower) {
			return &card
		}
	}

	return nil
}

// ParseIndex tries to parse a string as a 1-based index.
func ParseIndex(s string) (int, error) {
	return strconv.Atoi(s)
}
