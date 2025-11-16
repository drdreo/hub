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

type GameStatus string

const (
	GameStatusWaiting GameStatus = "waiting"
	GameStatusStarted GameStatus = "started"
	GameStatusEnded   GameStatus = "ended"
)

func (gs GameStatus) String() string {
	return string(gs)
}

type Game struct {
	dbService database.Database
}

type GameState struct {
	Ctx          context.Context
	RoomName     string            `json:"roomName"`
	Users        map[string]*User  `json:"users"`
	UserOrder    []string          `json:"userOrder"`
	Started      bool              `json:"started"`
	StartTime    time.Time         `json:"startTime"`
	GameStatus   GameStatus        `json:"gameStatus"`
	Stories      []*Story          `json:"stories"`
	FinishVotes  map[string]bool   `json:"finishVotes"`
	RestartVotes map[string]bool   `json:"restartVotes"`
	Config       models.RoomConfig `json:"config"`
}

func (s *GameState) ToMap() interfaces.M {
	users := make([]*models.UserDTO, 0, len(s.UserOrder))
	for _, uid := range s.UserOrder {
		if user, ok := s.Users[uid]; ok {
			users = append(users, user.ToDTO())
		}
	}

	return interfaces.M{
		"roomName":   s.RoomName,
		"users":      users,
		"started":    s.Started,
		"gameStatus": s.GameStatus.String(),
	}
}

func (g *Game) AddUser(clientId string, name string, state *GameState) {
	state.Users[clientId] = NewUser(clientId, name)
	state.UserOrder = append(state.UserOrder, clientId)
}

func (g *Game) GetUser(clientId string, state *GameState) *User {
	return state.Users[clientId]
}

func (g *Game) RemoveUser(clientID string, state *GameState) {
	user, exists := state.Users[clientID]
	if !exists {
		return
	}

	userName := user.Name
	delete(state.Users, clientID)

	for i, id := range state.UserOrder {
		if id == clientID {
			state.UserOrder = append(state.UserOrder[:i], state.UserOrder[i+1:]...)
			break
		}
	}

	log.Info().Str("user", userName).Str("room", state.RoomName).Msg("User removed from room")
}

// ReconnectUser handles reconnecting a user with a new client ID
// This patches all references to the old ID throughout the game state
func (g *Game) ReconnectUser(oldClientID, newClientID string, state *GameState) error {
	// Find the old user
	oldUser, ok := state.Users[oldClientID]
	if !ok {
		return errors.New("user not found")
	}

	// Update the user's ID and disconnected state
	oldUser.ID = newClientID
	oldUser.Disconnected = false

	// Move user to new ID in Users map
	delete(state.Users, oldClientID)
	state.Users[newClientID] = oldUser

	// Update UserOrder
	for i, uid := range state.UserOrder {
		if uid == oldClientID {
			state.UserOrder[i] = newClientID
			break
		}
	}

	// Update FinishVotes
	if voted, ok := state.FinishVotes[oldClientID]; ok {
		delete(state.FinishVotes, oldClientID)
		state.FinishVotes[newClientID] = voted
	}

	// Update RestartVotes
	if voted, ok := state.RestartVotes[oldClientID]; ok {
		delete(state.RestartVotes, oldClientID)
		state.RestartVotes[newClientID] = voted
	}

	// Update KickVotes for all users
	for _, user := range state.Users {
		for i, voterID := range user.KickVotes {
			if voterID == oldClientID {
				user.KickVotes[i] = newClientID
			}
		}
	}

	// Update Story owner
	for _, story := range state.Stories {
		if story.OwnerID == oldClientID {
			story.OwnerID = newClientID
			break
		}
	}

	log.Info().
		Str("user", oldUser.Name).
		Str("oldID", oldClientID).
		Str("newID", newClientID).
		Msg("User reconnected with new ID")

	return nil
}

func (g *Game) StartGame(state *GameState) {
	state.Started = true
	state.GameStatus = GameStatusStarted
	state.StartTime = time.Now()
	state.FinishVotes = make(map[string]bool)
	state.RestartVotes = make(map[string]bool)
	log.Info().Str("room", state.RoomName).Msg("Game started")
}

func (g *Game) GetStories(state *GameState) []models.StoryDTO {
	stories := make([]models.StoryDTO, 0, len(state.Stories))
	for _, story := range state.Stories {
		// Find the author name
		author := "Unknown"
		if user, ok := state.Users[story.OwnerID]; ok {
			author = user.Name
		}

		stories = append(stories, models.StoryDTO{
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

	if g.isUserStoryOwner(userID, state) {
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

	room.SetState(state)

	// Send story updates to all users who are story owners and have a queued story
	g.SendStoryUpdatesToOwners(state, room)
	g.SendUsersUpdate(state, room)

	return nil
}

// isUserStoryOwner checks if a user has created at least one story (submitted at least once)
func (g *Game) isUserStoryOwner(userID string, state *GameState) bool {
	for _, s := range state.Stories {
		if s.OwnerID == userID {
			return true
		}
	}
	return false
}

// SendStoryUpdatesToOwners sends story updates to all users who are story owners and have a queued story
// This prevents first-round stories from appearing before users submit their initial text,
// while ensuring all story owners with queued work are notified
func (g *Game) SendStoryUpdatesToOwners(state *GameState, room interfaces.Room) {
	// Iterate through all users in the room
	for _, userID := range state.UserOrder {
		g.SendStoryUpdate(userID, state, room)
	}
}

func (g *Game) SendStoryUpdate(userID string, state *GameState, room interfaces.Room) {
	user := state.Users[userID]
	if user == nil {
		return
	}

	log.Debug().Msg("send story update user exists")

	// Don't send story updates to users who haven't submitted yet
	if !g.isUserStoryOwner(userID, state) {
		log.Debug().Str("userId ", userID).Any("stories", state.Stories).Msg("send story update user is no owner")

		return
	}

	var storyData *models.StoryDTO

	story := user.GetCurrentStory()
	log.Debug().Any("story", story).Msg("Sending story update")

	if story != nil {
		// Find the author name
		author := "Unknown"
		if authorUser, ok := state.Users[story.OwnerID]; ok {
			author = authorUser.Name
		}

		storyData = &models.StoryDTO{
			Text:   story.GetLatestText(),
			Author: author,
		}
	}

	msg := protocol.NewSuccessResponse("story_update", storyData)
	room.SendTo(msg, userID)
}

func (g *Game) SendUsersUpdate(state *GameState, room interfaces.Room) {
	users := make([]models.UserDTO, 0, len(state.UserOrder))
	for _, uid := range state.UserOrder {
		if user, ok := state.Users[uid]; ok {
			users = append(users, models.UserDTO{
				ID:            user.ID,
				Name:          user.Name,
				Disconnected:  user.Disconnected,
				AFK:           user.AFK,
				KickVotes:     user.KickVotes,
				QueuedStories: len(user.StoryQueue),
			})
		}
	}

	msg := protocol.NewSuccessResponse("users_update", interfaces.M{"users": users})
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
	if state.FinishVotes[user.ID] {
		delete(state.FinishVotes, user.ID)
	} else {
		state.FinishVotes[user.ID] = true
	}

	// Send vote update
	votedIDs := make([]string, 0, len(state.FinishVotes))
	for id := range state.FinishVotes {
		votedIDs = append(votedIDs, id)
	}

	room.SetState(state)

	msg := protocol.NewSuccessResponse("finish_vote_update", interfaces.M{"votes": votedIDs})
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
	if state.RestartVotes[user.ID] {
		delete(state.RestartVotes, user.ID)
	} else {
		state.RestartVotes[user.ID] = true
	}

	// Send vote update
	votedIDs := make([]string, 0, len(state.RestartVotes))
	for id := range state.RestartVotes {
		votedIDs = append(votedIDs, id)
	}

	room.SetState(state)
	msg := protocol.NewSuccessResponse("restart_vote_update", interfaces.M{"votes": votedIDs})
	room.Broadcast(msg)

	// Check if all users voted
	if len(state.RestartVotes) >= len(state.Users) {
		g.RestartGame(state, room)
	}
}

func (g *Game) EndGame(state *GameState, room interfaces.Room) {
	state.GameStatus = GameStatusEnded
	room.SetState(state)

	log.Info().Str("room", state.RoomName).Msg("Game ended")

	stories := g.GetStories(state)
	msg := protocol.NewSuccessResponse("final_stories", interfaces.M{
		"stories": stories,
	})
	room.Broadcast(msg)

	msg = protocol.NewSuccessResponse("game_status", interfaces.M{
		"status": state.GameStatus.String(),
	})
	room.Broadcast(msg)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if g.dbService == nil {
		log.Warn().Str("room", state.RoomName).Msg("dbService is nil; stories not persisted")
		return
	}
	if err := g.dbService.StoreStories(ctx, state.RoomName, stories); err != nil {
		log.Error().Err(err).Msg("Failed to persist stories")
	}
}

func (g *Game) RestartGame(state *GameState, room interfaces.Room) {
	state.Started = false
	state.GameStatus = GameStatusWaiting
	state.Stories = make([]*Story, 0)
	state.FinishVotes = make(map[string]bool)
	state.RestartVotes = make(map[string]bool)

	for _, user := range state.Users {
		user.Reset()
	}

	room.SetState(state)

	g.SendUsersUpdate(state, room)

	msg := protocol.NewSuccessResponse("game_status", interfaces.M{
		"status": state.GameStatus.String(),
	})
	room.Broadcast(msg)

	msg = protocol.NewSuccessResponse("story_update", interfaces.M{
		"story": nil,
	})
	room.Broadcast(msg)

	msg = protocol.NewSuccessResponse("finish_vote_update", interfaces.M{"votes": state.FinishVotes})
	room.Broadcast(msg)
	msg = protocol.NewSuccessResponse("restart_vote_update", interfaces.M{"votes": state.RestartVotes})
	room.Broadcast(msg)

	log.Info().Str("room", state.RoomName).Msg("Game restarted")
}
