package tell_it

import (
	"context"
	"errors"
	"gameserver/games/tell_it/database"
	"gameserver/games/tell_it/models"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"github.com/rs/zerolog/log"
	"time"
)

type Game struct {
	dbService database.Database
}

type GameState struct {
	Ctx          context.Context
	RoomName     string            `json:"roomName"`
	Users        map[string]*User  `json:"users"`
	UserOrder    []string          `json:"userOrder"`
	Started      bool              `json:"started"`
	GameStatus   string            `json:"gameStatus"` // "waiting", "started", "ended"
	Stories      []*Story          `json:"stories"`
	FinishVotes  map[string]bool   `json:"finishVotes"`
	RestartVotes map[string]bool   `json:"restartVotes"`
	Config       models.RoomConfig `json:"config"`
	StartTime    time.Time         `json:"startTime"`
}

type User struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Disconnected bool     `json:"disconnected"`
	AFK          bool     `json:"afk"`
	KickVotes    []string `json:"kickVotes"`
	StoryQueue   []*Story `json:"storyQueue"`
}

type Story struct {
	ID      string   `json:"id"`
	OwnerID string   `json:"ownerId"`
	Texts   []string `json:"texts"`
}

func (s *GameState) ToMap() interfaces.M {
	users := make([]models.UserOverview, 0, len(s.UserOrder))
	for _, uid := range s.UserOrder {
		if user, ok := s.Users[uid]; ok {
			users = append(users, models.UserOverview{
				ID:            user.ID,
				Name:          user.Name,
				Disconnected:  user.Disconnected,
				AFK:           user.AFK,
				KickVotes:     user.KickVotes,
				QueuedStories: len(user.StoryQueue),
			})
		}
	}

	return interfaces.M{
		"roomName":   s.RoomName,
		"users":      users,
		"started":    s.Started,
		"gameStatus": s.GameStatus,
	}
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
	result := ""
	for i, text := range s.Texts {
		if i > 0 {
			result += ". "
		}
		result += text
	}
	return result
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

func (g *Game) AddUser(id string, name string, state *GameState) {
	state.Users[id] = NewUser(id, name)
	state.UserOrder = append(state.UserOrder, id)
}

func (g *Game) GetUser(id string, state *GameState) *User {
	return state.Users[id]
}

func (g *Game) RemoveUser(clientId string, room interfaces.Room) {
	state := room.State().(*GameState)
	userName := state.Users[clientId].Name
	delete(state.Users, clientId)

	// Remove from user order
	for i, id := range state.UserOrder {
		if id == clientId {
			state.UserOrder = append(state.UserOrder[:i], state.UserOrder[i+1:]...)
			break
		}
	}

	log.Info().Str("user", userName).Str("room", room.ID()).Msg("User removed from room")
}

func (g *Game) StartGame(state *GameState) {
	state.Started = true
	state.GameStatus = "started"
	state.StartTime = time.Now()
	state.FinishVotes = make(map[string]bool)
	state.RestartVotes = make(map[string]bool)
	log.Info().Str("room", state.RoomName).Msg("Game started")
}

func (g *Game) GetStories(state *GameState) []models.StoryData {
	stories := make([]models.StoryData, 0, len(state.Stories))
	for _, story := range state.Stories {
		// Find the author name
		author := "Unknown"
		if user, ok := state.Users[story.OwnerID]; ok {
			author = user.Name
		}

		stories = append(stories, models.StoryData{
			Text:   story.Serialize(),
			Author: author,
		})
	}
	return stories
}

func (g *Game) SubmitText(userID string, text string, state *GameState, room interfaces.Room) error {
	// Find new user to continue
	currentIndex := -1
	for i, uid := range state.UserOrder {
		if uid == userID {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return errors.New("user not found")
	}

	nextIndex := (currentIndex + 1) % len(state.UserOrder)
	nextUserID := state.UserOrder[nextIndex]
	nextUser := state.Users[nextUserID]

	var story *Story

	// Check if this user already owns a story
	user := state.Users[userID]
	isOwner := false
	for _, s := range state.Stories {
		if s.OwnerID == userID {
			isOwner = true
			break
		}
	}

	if isOwner {
		// Try to dequeue the user's current story
		var err error
		story, err = user.DequeueStory()
		if err != nil {
			return errors.New("can't wait - no story to continue")
		}
	} else {
		// Create a new story
		story = NewStory(userID)
		state.Stories = append(state.Stories, story)
	}

	story.AddText(text)
	nextUser.EnqueueStory(story)

	// Send story update to the next user
	g.SendStoryUpdate(nextUserID, state, room)
	g.SendUsersUpdate(state, room)

	return nil
}

func (g *Game) SendStoryUpdate(userID string, state *GameState, room interfaces.Room) {
	user := state.Users[userID]
	if user == nil {
		return
	}

	story := user.GetCurrentStory()
	if story == nil {
		return
	}

	// Find the author name
	author := "Unknown"
	if authorUser, ok := state.Users[story.OwnerID]; ok {
		author = authorUser.Name
	}

	storyData := models.StoryData{
		Text:   story.GetLatestText(),
		Author: author,
	}

	msg := protocol.NewSuccessResponse("story_update", storyData)

	// Send to specific user
	clients := room.Clients()
	if client, ok := clients[userID]; ok {
		client.Send(msg)
	}
}

func (g *Game) SendUsersUpdate(state *GameState, room interfaces.Room) {
	users := make([]models.UserOverview, 0, len(state.UserOrder))
	for _, uid := range state.UserOrder {
		if user, ok := state.Users[uid]; ok {
			users = append(users, models.UserOverview{
				ID:            user.ID,
				Name:          user.Name,
				Disconnected:  user.Disconnected,
				AFK:           user.AFK,
				KickVotes:     user.KickVotes,
				QueuedStories: len(user.StoryQueue),
			})
		}
	}

	msg := protocol.NewSuccessResponse("users_update", map[string]interface{}{"users": users})
	room.Broadcast(msg)
}

func (g *Game) VoteFinish(userID string, state *GameState, room interfaces.Room) {
	user := state.Users[userID]
	if user == nil {
		log.Error().Str("userID", userID).Msg("User not found for finish vote")
		return
	}

	log.Info().Str("user", user.Name).Msg("User voted to finish the game")

	// Toggle vote
	if state.FinishVotes[user.Name] {
		delete(state.FinishVotes, user.Name)
	} else {
		state.FinishVotes[user.Name] = true
	}

	// Send vote update
	votes := make([]string, 0, len(state.FinishVotes))
	for name := range state.FinishVotes {
		votes = append(votes, name)
	}

	msg := protocol.NewSuccessResponse("finish_vote_update", map[string]interface{}{"votes": votes})
	room.Broadcast(msg)

	// Check if all users voted
	if len(state.FinishVotes) >= len(state.Users) {
		g.EndGame(state, room)
	}
}

func (g *Game) VoteRestart(userID string, state *GameState, room interfaces.Room) {
	user := state.Users[userID]
	if user == nil {
		log.Error().Str("userID", userID).Msg("User not found for restart vote")
		return
	}

	log.Info().Str("user", user.Name).Msg("User voted to restart the game")

	// Toggle vote
	if state.RestartVotes[user.Name] {
		delete(state.RestartVotes, user.Name)
	} else {
		state.RestartVotes[user.Name] = true
	}

	// Check if all users voted
	if len(state.RestartVotes) >= len(state.Users) {
		g.RestartGame(state, room)
	}
}

func (g *Game) EndGame(state *GameState, room interfaces.Room) {
	state.GameStatus = "ended"
	log.Info().Str("room", state.RoomName).Msg("Game ended")

	// Send final stories
	stories := g.GetStories(state)
	msg := protocol.NewSuccessResponse("final_stories", map[string]interface{}{
		"stories": stories,
	})
	room.Broadcast(msg)

	// Persist stories to database
	if err := g.dbService.StoreStories(state.Ctx, state.RoomName, stories); err != nil {
		log.Error().Err(err).Msg("Failed to persist stories")
	}
}

func (g *Game) RestartGame(state *GameState, room interfaces.Room) {
	state.Started = false
	state.GameStatus = "waiting"
	state.Stories = make([]*Story, 0)
	state.FinishVotes = make(map[string]bool)
	state.RestartVotes = make(map[string]bool)

	// Reset all users
	for _, user := range state.Users {
		user.Reset()
	}

	g.SendUsersUpdate(state, room)

	// Send reset story update to all users
	msg := protocol.NewSuccessResponse("story_update", map[string]interface{}{
		"story": nil,
	})
	room.Broadcast(msg)

	log.Info().Str("room", state.RoomName).Msg("Game restarted")
}
