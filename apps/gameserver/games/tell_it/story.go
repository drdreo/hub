package tell_it

import (
	"gameserver/games/tell_it/models"
	"strings"
)

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

func (s *Story) GetStats() models.StoryStats {
	words := 0
	for _, text := range s.Texts {
		words += len(strings.Fields(text))
	}

	return models.StoryStats{
		Turns:          len(s.Texts),
		Words:          words,
		AvgReadingTime: calculateWordsPerSecond(words),
	}
}

func (s *Story) ToDTO(author string, includeAll bool) *models.StoryDTO {
	text := ""
	if includeAll {
		text = s.Serialize()
	} else {
		text = s.GetLatestText()
	}

    return &models.StoryDTO{
        Text:   text,
        Author: author,
        Stats:  s.GetStats(),
    }
}

func calculateWordsPerSecond(words int) float64 {
	wpm := float64(words) / 200
	return wpm * 60
}
