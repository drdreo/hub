package tell_it

import "strings"

// Story represents a single story in the game
type Story struct {
	ID      string   `json:"id"`
	OwnerID string   `json:"ownerId"`
	Texts   []string `json:"texts"`
}

func NewStory(ownerID string) *Story {
	return &Story{
		OwnerID: ownerID,
		Texts:   make([]string, 0),
	}
}

func (s *Story) AddText(text string) {
	s.Texts = append(s.Texts, text)
}

func (s *Story) GetLatestText() string {
	if len(s.Texts) == 0 {
		return ""
	}
	return s.Texts[len(s.Texts)-1]
}

func (s *Story) Serialize() string {
	return strings.Join(s.Texts, ". ")
}
