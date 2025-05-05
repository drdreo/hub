package testicles

import (
	"context"
	"fmt"
	"gameserver/internal/client"
	"gameserver/internal/game"
	"gameserver/internal/interfaces"
	"gameserver/internal/room"
	"gameserver/internal/router"
	"testing"
)

// TestHelper provides a test setup for game integration tests
type TestHelper struct {
	Ctx           context.Context
	Registry      interfaces.GameRegistry
	ClientManager interfaces.ClientManager
	RoomManager   interfaces.RoomManager
	Router        *router.Router
	Clients       map[string]*client.ClientMock
	RoomID        string
	t             *testing.T
}

// NewTestHelper creates a new test helper for game integration tests
func NewTestHelper(t *testing.T) *TestHelper {
	testCtx := context.Background()

	// Set up the complete system with real components
	registry := game.NewRegistry()
	clientManager := client.NewManager()
	roomManager := room.NewRoomManager(registry, nil)
	testRouter := router.NewRouter(testCtx, clientManager, roomManager, registry, nil)

	return &TestHelper{
		Ctx:           testCtx,
		Registry:      registry,
		ClientManager: clientManager,
		RoomManager:   roomManager,
		Router:        testRouter,
		Clients:       make(map[string]*client.ClientMock),
		t:             t,
	}
}

// RegisterGame registers a game with the registry
func (th *TestHelper) RegisterGame(g interfaces.Game) {
	th.Registry.RegisterGame(g)
}

// CreateClient creates a new mock client with the given ID and adds it to the clients map
func (th *TestHelper) CreateClient(id string) *client.ClientMock {
	c := client.NewClientMock(id)
	th.Clients[id] = c
	return c
}

// CreateRoom creates a new room for the given game type and returns the room ID
func (th *TestHelper) CreateRoom(client *client.ClientMock, gameType, playerName string) string {
	joinRoomMsg := fmt.Sprintf(`{"type":"join_room","data":{"gameType":"%s","playerName":"%s"}}`, gameType, playerName)
	th.Router.HandleMessage(client, []byte(joinRoomMsg))

	messages := client.GetSentMessages()
	if len(messages) == 0 {
		th.t.Fatalf("No messages received after room creation")
	}

	// Extract room ID from response
	createResponse := messages[len(messages)-1]
	roomId, ok := ExtractJoinRoomResponseData(createResponse)
	if !ok {
		th.t.Fatalf("invalid join response")
	}

	th.RoomID = roomId

	client.ClearMessages()

	return th.RoomID
}

// JoinRoom makes a client join an existing room
func (th *TestHelper) JoinRoom(client interfaces.Client, roomID, playerName string) {
	joinMessage := fmt.Sprintf(`{"type":"join_room","data":{"roomId":"%s", "playerName":"%s"}}`, roomID, playerName)
	th.Router.HandleMessage(client, []byte(joinMessage))
}

// GetRoom retrieves the room with the current room ID
func (th *TestHelper) GetRoom() (interfaces.Room, error) {
	return th.RoomManager.GetRoom(th.RoomID)
}

// ClearAllMessages clears messages for all registered clients
func (th *TestHelper) ClearAllMessages() {
	for _, c := range th.Clients {
		c.ClearMessages()
	}
}

// SendMessage sends a message from a client to the router
func (th *TestHelper) SendMessage(clientID string, msgType string, data interface{}) {
	c, exists := th.Clients[clientID]
	if !exists {
		th.t.Fatalf("Client with ID %s not found", clientID)
	}

	var msgBytes []byte

	if rawData, ok := data.([]byte); ok {
		msgBytes = rawData
	} else {
		msgBytes = CreateGameMessage(msgType, data)
	}

	th.Router.HandleMessage(c, msgBytes)
}

// VerifyMessageReceived checks if a client received a message of the specified type
func (th *TestHelper) VerifyMessageReceived(clientID, msgType string) bool {
	c, exists := th.Clients[clientID]
	if !exists {
		th.t.Fatalf("Client with ID %s not found", clientID)
	}

	messages := c.GetSentMessages()
	for _, msg := range messages {
		if msg.Type == msgType {
			return true
		}
	}

	return false
}

// AssertMessageReceived asserts that a client received a message of the specified type
func (th *TestHelper) AssertMessageReceived(clientID, msgType string) {
	if !th.VerifyMessageReceived(clientID, msgType) {
		th.t.Errorf("Client %s did not receive a message of type %s", clientID, msgType)
	}
}

// AssertClientsReceivedMessages asserts that all specified clients received messages
func (th *TestHelper) AssertClientsReceivedMessages(clientIDs []string) {
	for _, id := range clientIDs {
		c, exists := th.Clients[id]
		if !exists {
			th.t.Fatalf("Client with ID %s not found", id)
		}

		messages := c.GetSentMessages()
		if len(messages) == 0 {
			th.t.Errorf("Client %s didn't receive any messages", id)
		}
	}
}

// SetupGameRoom sets up a game room with two players
func (th *TestHelper) SetupGameRoom(gameType string, amountPlayers int) []string {
	playerIds := make([]string, amountPlayers)

	for idx := range amountPlayers {
		clientId := fmt.Sprintf("player-%d", idx)
		c := th.CreateClient(clientId)
		if idx == 0 {
			th.CreateRoom(c, gameType, clientId)
		} else {
			th.JoinRoom(c, th.RoomID, clientId)
		}
		playerIds[idx] = clientId
	}

	th.AssertClientsReceivedMessages(playerIds)

	th.ClearAllMessages()

	return playerIds
}
