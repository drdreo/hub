package tell_it

import (
	"errors"
	"gameserver/games/tell_it/models"
)

type User struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Disconnected bool     `json:"disconnected"`
	AFK          bool     `json:"afk"`
	KickVotes    []string `json:"kickVotes"`
	StoryQueue   []*Story `json:"storyQueue"`
}

func NewUser(id, name string) *User {
	return &User{
		ID:           id,
		Name:         name,
		Disconnected: false,
		AFK:          false,
		KickVotes:    make([]string, 0),
		StoryQueue:   make([]*Story, 0),
	}
}

func (u *User) ToDTO() *models.UserDTO {
	return &models.UserDTO{
		ID:            u.ID,
		Name:          u.Name,
		Disconnected:  u.Disconnected,
		AFK:           u.AFK,
		KickVotes:     u.KickVotes,
		QueuedStories: len(u.StoryQueue),
	}
}

func (u *User) EnqueueStory(story *Story) {
	u.StoryQueue = append(u.StoryQueue, story)
}

func (u *User) DequeueStory() (*Story, error) {
	if len(u.StoryQueue) == 0 {
		return nil, errors.New("no stories in queue")
	}
	story := u.StoryQueue[0]
	u.StoryQueue = u.StoryQueue[1:]
	return story, nil
}

func (u *User) GetCurrentStory() *Story {
	if len(u.StoryQueue) == 0 {
		return nil
	}
	return u.StoryQueue[0]
}

func (u *User) Reset() {
	u.AFK = false
	u.KickVotes = make([]string, 0)
	u.StoryQueue = make([]*Story, 0)
}
