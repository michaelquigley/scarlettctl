package scarlettctl

import (
	"fmt"
	"strings"
)

// OpenCard opens an ALSA control connection to the specified card number
func OpenCard(cardNum int) (*Card, error) {
	handle, err := openCard(cardNum)
	if err != nil {
		return nil, err
	}

	name, err := getCardInfo(cardNum)
	if err != nil {
		closeCard(handle)
		return nil, err
	}

	return &Card{
		Number: cardNum,
		Name:   name,
		handle: handle,
	}, nil
}

// Close closes the connection to the card
func (c *Card) Close() error {
	if c.handle == nil {
		return nil
	}
	return closeCard(c.handle)
}

// String returns a string representation of the card
func (c *Card) String() string {
	return fmt.Sprintf("Card %d: %s", c.Number, c.Name)
}

// ListCards returns a list of all available ALSA cards
// It filters to only include Focusrite Scarlett devices
func ListCards() ([]*Card, error) {
	cards := make([]*Card, 0)

	// try card numbers 0-7 (typical range)
	for i := 0; i < 8; i++ {
		name, err := getCardInfo(i)
		if err != nil {
			continue // card doesn't exist or can't be accessed
		}

		// filter for Scarlett devices
		nameLower := strings.ToLower(name)
		if strings.Contains(nameLower, "scarlett") ||
		   strings.Contains(nameLower, "focusrite") ||
		   strings.Contains(nameLower, "vocaster") ||
		   strings.Contains(nameLower, "clarett") {
			cards = append(cards, &Card{
				Number: i,
				Name:   name,
			})
		}
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("no Focusrite Scarlett/Vocaster/Clarett devices found")
	}

	return cards, nil
}

// FindCard finds a card by number or name substring
func FindCard(identifier string) (*Card, error) {
	cards, err := ListCards()
	if err != nil {
		return nil, err
	}

	// try parsing as card number
	var cardNum int
	if _, err := fmt.Sscanf(identifier, "%d", &cardNum); err == nil {
		for _, card := range cards {
			if card.Number == cardNum {
				return OpenCard(card.Number)
			}
		}
		return nil, fmt.Errorf("card %d not found", cardNum)
	}

	// try matching by name substring
	identifierLower := strings.ToLower(identifier)
	for _, card := range cards {
		if strings.Contains(strings.ToLower(card.Name), identifierLower) {
			return OpenCard(card.Number)
		}
	}

	return nil, fmt.Errorf("no card matching '%s' found", identifier)
}

// IsScarlett checks if this card is a supported Scarlett device
func (c *Card) IsScarlett() bool {
	nameLower := strings.ToLower(c.Name)
	return strings.Contains(nameLower, "scarlett") ||
	       strings.Contains(nameLower, "vocaster") ||
	       strings.Contains(nameLower, "clarett")
}

// GetPollFds returns the file descriptors to poll for events
func (c *Card) GetPollFds() []int {
	if c.handle == nil {
		return nil
	}
	return c.handle.pollFds
}
