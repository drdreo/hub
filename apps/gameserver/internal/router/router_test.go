package router

import (
	"encoding/json"
	testgame "gameserver/games/test"
	"gameserver/internal/client"
	"gameserver/internal/game"
	"gameserver/internal/protocol"
	"gameserver/internal/room"
	"gameserver/internal/session"
	"testing"
)

func TestRouter(t *testing.T) {
	session.InitGlobalStore(2)

	registry := game.NewRegistry()
	testgame.RegisterTestGame(registry)
	roomManager := room.NewRoomManager(registry)
	router := NewRouter(roomManager, registry)

	t.Run("invalid message format", func(t *testing.T) {
		client := client.NewClientMock("test1")
		router.HandleMessage(client, []byte("invalid json"))

		messages := client.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 error message, got %d", len(messages))
		}
		if messages[0].Type != "error" {
			t.Errorf("expected error message type, got %s", messages[0].Type)
		}
	})

	t.Run("create room with invalid options", func(t *testing.T) {
		client := client.NewClientMock("test1")
		msg := protocol.Message{
			Type: "create_room",
			Data: json.RawMessage(`{"invalid": "json"}`),
		}
		msgData, _ := json.Marshal(msg)
		router.HandleMessage(client, msgData)

		messages := client.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 error message, got %d", len(messages))
		}
		response := messages[0]
		if response.Success != false {
			t.Errorf("expected success to be false, got %t", response.Success)
		}
		if response.Type != "error" {
			t.Errorf("expected error message type, got %s", messages[0].Type)
		}
	})

	t.Run("create room with missing game type", func(t *testing.T) {
		client := client.NewClientMock("test1")
		msg := protocol.Message{
			Type: "join_room",
			Data: json.RawMessage(`{}`),
		}
		msgData, _ := json.Marshal(msg)
		router.HandleMessage(client, msgData)

		messages := client.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 error message, got %d", len(messages))
		}
		response := messages[0]
		if response.Success != false {
			t.Errorf("expected success to be false, got %t", response.Success)
		}
		if response.Type != "join_room_result" {
			t.Errorf("expected error message type, got %s", messages[0].Type)
		}
	})

	t.Run("create and join room flow", func(t *testing.T) {
		client1 := client.NewClientMock("client1")
		msg := protocol.Message{
			Type: "join_room",
			Data: json.RawMessage(`{"gameType": "testGame", "playerName": "tester-1"}`),
		}
		msgData, _ := json.Marshal(msg)
		router.HandleMessage(client1, msgData)

		// Verify create room response
		messages := client1.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Type != "join_room_result" {
			t.Errorf("expected join_room_result, got %s", messages[0].Type)
		}

		// Get room ID from response
		response := messages[0].Data.(map[string]interface{})
		roomID := response["roomId"].(string)

		// Create second client and join room
		client2 := client.NewClientMock("client2")
		msg = protocol.Message{
			Type: "join_room",
			Data: json.RawMessage(`{"roomId": "` + roomID + `", "playerName": "tester-2"}`),
		}
		msgData, _ = json.Marshal(msg)
		router.HandleMessage(client2, msgData)

		// Verify join room response
		messages = client2.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Type != "join_room_result" {
			t.Errorf("expected join_room_result, got %s", messages[0].Type)
		}
	})

	t.Run("leave room flow", func(t *testing.T) {
		client := client.NewClientMock("client3")
		msg := protocol.Message{
			Type: "create_room",
			Data: json.RawMessage(`{"gameType": "testGame"}`),
		}
		msgData, _ := json.Marshal(msg)
		router.HandleMessage(client, msgData)

		client.ClearMessages()

		msg = protocol.Message{
			Type: "leave_room",
		}
		msgData, _ = json.Marshal(msg)
		router.HandleMessage(client, msgData)

		// Verify leave room response
		messages := client.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Type != "leave_room_result" {
			t.Errorf("expected leave_room_result, got %s", messages[0].Type)
		}
	})

	// TODO: add success reconnect flow
	
	t.Run("reconnect flow of foreign client should fail", func(t *testing.T) {
		client1 := client.NewClientMock("client4")
		msg := protocol.Message{
			Type: "join_room",
			Data: json.RawMessage(`{"gameType": "testGame", "playerName": "tester-1"}`),
		}
		msgData, _ := json.Marshal(msg)
		router.HandleMessage(client1, msgData)

		// Get room ID from response
		response := client1.GetSentMessages()[0].Data.(map[string]interface{})
		roomID := response["roomId"].(string)

		// Create new client for reconnection
		client2 := client.NewClientMock("client5")
		msg = protocol.Message{
			Type: "reconnect",
			Data: json.RawMessage(`{"clientId": "client4", "roomId": "` + roomID + `"}`),
		}
		msgData, _ = json.Marshal(msg)
		router.HandleMessage(client2, msgData)

		// Verify reconnect response
		messages := client2.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Success != false {
			t.Errorf("expected reconnect_result sucess to be false, got %s", messages[0].Type)
		}
		if messages[0].Type != "reconnect_result" {
			t.Errorf("expected reconnect_result, got %s", messages[0].Type)
		}
	})
}
