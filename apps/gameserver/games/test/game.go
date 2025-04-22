package testgame

import (
	"encoding/json"
	"gameserver/internal/interfaces"
)

type TestGame struct{}

type PlayerInfo struct {
	Name string `json:"name"`
}

type GameState struct {
	Players map[string]PlayerInfo `json:"players"`
}

func (g *TestGame) Type() string {
	return "testGame"
}

func (g *TestGame) HandleMessage(client interfaces.Client, room interfaces.Room, msgType string, payload []byte) error {
	return nil
}

func (g *TestGame) InitializeRoom(room interfaces.Room, options json.RawMessage) error {
	state := GameState{
		Players: make(map[string]PlayerInfo),
	}

	room.SetState(state)
	return nil
}

func (g *TestGame) OnClientJoin(client interfaces.Client, room interfaces.Room, _ interfaces.CreateRoomOptions) {
}

func (g *TestGame) OnBotAdd(client interfaces.Client, room interfaces.Room, reg interfaces.GameRegistry) (interfaces.Client, error) {
	bot := NewBot("bot-1", g, reg)
	return bot.BotClient, nil
}

func (g *TestGame) OnClientLeave(client interfaces.Client, room interfaces.Room) {
}

func (g *TestGame) OnClientReconnect(client interfaces.Client, room interfaces.Room, oldClientId string) {
}

func NewTestGame() *TestGame {
	return &TestGame{}
}

func RegisterTestGame(r interfaces.GameRegistry) {
	g := NewTestGame()
	r.RegisterGame(g)
}
