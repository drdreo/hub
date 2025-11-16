package tell_it

import (
	"context"
	"encoding/json"
	"errors"
	"gameserver/games/tell_it/database"
	"gameserver/games/tell_it/models"
	"gameserver/internal/interfaces"
	"gameserver/internal/protocol"
	"github.com/rs/zerolog/log"
	"time"
)

type GameConfig struct {
	Stage interfaces.Environment
}

func NewGame(dbService database.Database) *Game {
	return &Game{
		dbService: dbService,
	}
}

func RegisterGame(ctx context.Context, r interfaces.GameRegistry, config GameConfig) error {
	dbInitCtx, dbInitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbInitCancel()

	dbFactory := database.NewDatabaseFactory(config.Stage)
	dbService, err := dbFactory.CreateDatabaseService(dbInitCtx)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize database service for tell-it")
		return err
	}

	g := NewGame(dbService)
	r.RegisterGame(g)
	return nil
}

// Type returns the game type
func (g *Game) Type() string {
	return "tellit"
}

// InitializeRoom sets up a new room with the initial game state
func (g *Game) InitializeRoom(ctx context.Context, room interfaces.Room, options json.RawMessage) error {
	// Parse room config if provided
	config := models.RoomConfig{
		SpectatorsAllowed: true,
		IsPublic:          true,
		MinUsers:          2,
		MaxUsers:          30,
		AFKDelay:          30000, // 30 seconds
	}

	if options != nil {
		if err := json.Unmarshal(options, &config); err != nil {
			log.Warn().Err(err).Msg("Failed to parse room config, using defaults")
		}
	}

	state := GameState{
		RoomName:     room.ID(),
		Users:        make(map[string]*User),
		UserOrder:    make([]string, 0),
		Started:      false,
		GameStatus:   GameStatusWaiting,
		Stories:      make([]*Story, 0),
		FinishVotes:  make(map[string]bool),
		RestartVotes: make(map[string]bool),
		Config:       config,
	}

	room.SetState(&state)
	log.Info().Str("room", room.ID()).Msg("Tell-It room initialized")
	return nil
}

func (g *Game) OnClientJoin(client interfaces.Client, room interfaces.Room, options interfaces.CreateRoomOptions) {
	state := room.State().(*GameState)

	// Add user to the game
	userName := options.PlayerName

	g.AddUser(client.ID(), userName, state)

	log.Info().Str("user", userName).Str("room", room.ID()).Msg("User joined tell-it room")

	room.SetState(state)
	// Broadcast updated user list to all clients
	g.SendUsersUpdate(state, room)
}

func (g *Game) OnClientLeave(client interfaces.Client, room interfaces.Room) {
	state := room.State().(*GameState)

	user := g.GetUser(client.ID(), state)
	if user != nil {
		log.Info().Str("user", user.Name).Str("room", room.ID()).Msg("User left tell-it room")

		// Mark as disconnected instead of removing immediately
		user.Disconnected = true
		room.SetState(state)

		g.SendUsersUpdate(state, room)
	}
}

// OnClientReconnect handles when a client reconnects to the game
func (g *Game) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientId string) error {
	state := room.State().(*GameState)

	if err := g.ReconnectUser(oldClientId, client.ID(), state); err != nil {
		log.Error().Err(err).Str("oldID", oldClientId).Str("newID", client.ID()).Msg("Failed to reconnect user")
		return err
	}

	room.SetState(state)

	g.handleRequestUpdate(client, state, room)
	g.SendUsersUpdate(state, room)

	return nil
}

// OnBotAdd handles adding a bot to the game (not supported for tell-it)
func (g *Game) OnBotAdd(client interfaces.Client, room interfaces.Room, registry interfaces.GameRegistry) (interfaces.Client, string, error) {
	return nil, "", errors.New("bots are not supported for tell-it game")
}

func (g *Game) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, data []byte) error {
	state := room.State().(*GameState)

	switch msgType {
	case "start":
		g.handleStart(client, state, room)
	case "submit_text":
		g.handleSubmitText(client, state, room, data)
	case "vote_finish":
		g.handleVoteFinish(client, state, room)
	case "vote_restart":
		g.handleVoteRestart(client, state, room)
	case "vote_kick":
		g.handleVoteKick(client, state, room, data)
	case "request_stories":
		g.handleRequestStories(client, state, room)
	case "request_update":
		g.handleRequestUpdate(client, state, room)
	default:
		log.Warn().Str("type", msgType).Msg("Unknown message type for tell-it")
	}
	return nil
}

func (g *Game) handleStart(client interfaces.Client, state *GameState, room interfaces.Room) {
	if state.Started {
		log.Warn().Str("room", room.ID()).Msg("Game already started")
		return
	}

	if len(state.Users) < state.Config.MinUsers {
		log.Warn().Str("room", room.ID()).Int("users", len(state.Users)).Msg("Not enough users to start")
		return
	}

	g.StartGame(state)

	room.SetState(state)

	// Broadcast game status update
	msg := protocol.NewSuccessResponse("game_status", interfaces.M{
		"status": state.GameStatus.String(),
	})
	room.Broadcast(msg)
}

func (g *Game) handleSubmitText(client interfaces.Client, state *GameState, room interfaces.Room, payload json.RawMessage) {
	var data struct {
		Text string `json:"text"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		log.Error().Err(err).Msg("Failed to parse submit_text payload")
		return
	}

	if err := g.SubmitText(client.ID(), data.Text, state, room); err != nil {
		log.Error().Err(err).Str("user", client.ID()).Msg("Failed to submit text")
		// Send error back to client
		errMsg := protocol.NewErrorResponse("submit_text", err.Error())
		client.Send(errMsg)
	}
}

func (g *Game) handleVoteFinish(client interfaces.Client, state *GameState, room interfaces.Room) {
	g.VoteFinish(client.ID(), state, room)
}

func (g *Game) handleVoteRestart(client interfaces.Client, state *GameState, room interfaces.Room) {
	g.VoteRestart(client.ID(), state, room)
}

func (g *Game) handleVoteKick(client interfaces.Client, state *GameState, room interfaces.Room, payload json.RawMessage) {
	var data struct {
		KickUserID string `json:"kickUserID"`
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		log.Error().Err(err).Msg("Failed to parse vote_kick payload")
		return
	}

	user := g.GetUser(client.ID(), state)
	targetUser := g.GetUser(data.KickUserID, state)

	if user == nil || targetUser == nil {
		log.Error().Msg("User not found for kick vote")
		return
	}

	// Toggle kick vote
	alreadyVoted := false
	for i, voterID := range targetUser.KickVotes {
		if voterID == client.ID() {
			targetUser.KickVotes = append(targetUser.KickVotes[:i], targetUser.KickVotes[i+1:]...)
			alreadyVoted = true
			break
		}
	}

	if !alreadyVoted {
		targetUser.KickVotes = append(targetUser.KickVotes, client.ID())
	}

	log.Info().Str("voter", user.Name).Str("target", targetUser.Name).Int("votes", len(targetUser.KickVotes)).Msg("Kick vote")

	// Check if majority voted to kick (>50%)
	if len(targetUser.KickVotes) > len(state.Users)/2 {
		log.Info().Str("user", targetUser.Name).Msg("User kicked by vote")

		g.RemoveUser(data.KickUserID, state)

		kickMsg := protocol.NewSuccessResponse("user_kicked", interfaces.M{
			"kickedUser": targetUser.Name,
		})
		room.Broadcast(kickMsg)

		// Disconnect the client
		clients := room.Clients()
		if kickClient, ok := clients[data.KickUserID]; ok {
			kickClient.Close()
		}
	}

	g.SendUsersUpdate(state, room)
}

func (g *Game) handleRequestStories(client interfaces.Client, state *GameState, room interfaces.Room) {
	stories := g.GetStories(state)
	msg := protocol.NewSuccessResponse("final_stories", interfaces.M{
		"stories": stories,
	})
	client.Send(msg)
}

func (g *Game) handleRequestUpdate(client interfaces.Client, state *GameState, room interfaces.Room) {
	// Send current game state to the client
	g.SendUsersUpdate(state, room)

	// Send game status
	msg := protocol.NewSuccessResponse("game_status", interfaces.M{
		"status": state.GameStatus.String(),
	})
	client.Send(msg)

	if state.GameStatus == GameStatusEnded {
		g.handleRequestStories(client, state, room)
	} else if state.GameStatus == GameStatusStarted {
		// If user has a story, send it
		user := g.GetUser(client.ID(), state)
		if user != nil {
			g.SendStoryUpdate(user.ID, state, room)
		}
	}

}
