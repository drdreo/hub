package router

import (
	"context"
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

	testCtx := context.Background()
	registry := game.NewRegistry()
	testgame.RegisterTestGame(registry)
	clientManager := client.NewManager()
	roomManager := room.NewRoomManager(registry)
	router := NewRouter(testCtx, clientManager, roomManager, registry)

	t.Run("invalid message format", func(t *testing.T) {
		client1 := client.NewClientMock("test1")
		router.HandleMessage(client1, []byte("invalid json"))

		messages := client1.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 error message, got %d", len(messages))
		}
		if messages[0].Type != "error" {
			t.Errorf("expected error message type, got %s", messages[0].Type)
		}
	})

	t.Run("create room with invalid options", func(t *testing.T) {
		client1 := client.NewClientMock("test1")
		msgData := CreateMessage("create_room", map[string]interface{}{
			"invalid": "json",
		})
		router.HandleMessage(client1, msgData)

		messages := client1.GetSentMessages()
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
		client1 := client.NewClientMock("test1")
		msgData := CreateMessage("join_room", nil)
		router.HandleMessage(client1, msgData)

		messages := client1.GetSentMessages()
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
		msg := CreateMessage("join_room", map[string]interface{}{
			"gameType":   "testGame",
			"playerName": "tester-1",
		})
		router.HandleMessage(client1, msg)
		// Verify create room response
		messages := client1.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Type != "join_room_result" {
			t.Errorf("expected join_room_result, got %s", messages[0].Type)
		}

		// Get room ID from response
		response := messages[0].Data.(*JoinResponse)
		roomID := response.RoomID

		// Create second client and join room
		client2 := client.NewClientMock("client2")
		msg = CreateMessage("join_room", map[string]interface{}{
			"roomId":     roomID,
			"playerName": "tester-2",
		})
		router.HandleMessage(client2, msg)

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
		client1 := client.NewClientMock("client3")
		msg := CreateMessage("create_room", map[string]interface{}{
			"gameType": "testGame",
		})
		router.HandleMessage(client1, msg)

		client1.ClearMessages()

		msg = CreateMessage("leave_room", nil)
		router.HandleMessage(client1, msg)

		// Verify leave room response
		messages := client1.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}
		if messages[0].Type != "leave_room_result" {
			t.Errorf("expected leave_room_result, got %s", messages[0].Type)
		}
	})

	t.Run("successful reconnect flow", func(t *testing.T) {
		client1 := client.NewClientMock("client6")

		msg := CreateMessage("join_room", map[string]interface{}{
			"gameType":   "testGame",
			"playerName": "tester-1",
		})
		router.HandleMessage(client1, msg)

		// Get client1's response to extract room ID
		joinResponse := client1.GetSentMessages()[0]
		respData := joinResponse.Data.(*JoinResponse)
		roomID := respData.RoomID
		client1ID := respData.ClientID

		// Simulate client1 closing its connection, which triggers session storage
		client1.Close()

		client1.ClearMessages()

		// Create a new client for reconnection
		client2 := client.NewClientMock("client7")
		reconnectMsg := CreateMessage("reconnect", map[string]interface{}{
			"clientId": client1ID,
			"roomID":   roomID,
		})
		router.HandleMessage(client2, reconnectMsg)

		// Verify reconnect response
		messages := client2.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 message, got %d", len(messages))
		}

		reconnectResponse := messages[0]
		if reconnectResponse.Type != "reconnect_result" {
			t.Errorf("expected reconnect_result, got %s", reconnectResponse.Type)
		}
		if !reconnectResponse.Success {
			t.Errorf("reconnect should succeed but got failure: %v", reconnectResponse.Data)
		}

		// Verify the response data contains expected fields
		reconResp := reconnectResponse.Data.(*ReconnectResponse)
		if reconResp.RoomID != roomID {
			t.Errorf("reconnect response has wrong roomId, got %v, expected %s", reconResp.RoomID, roomID)
		}
		if reconResp.GameType != "testGame" {
			t.Errorf("reconnect response has wrong gameType, got %v, expected %s", reconResp.GameType, "testGame")
		}
	})

	t.Run("reconnect flow of foreign client should fail", func(t *testing.T) {
		client1 := client.NewClientMock("client4")
		msg := CreateMessage("join_room", map[string]interface{}{
			"gameType":   "testGame",
			"playerName": "tester-1",
		})

		router.HandleMessage(client1, msg)

		// Get room ID from response
		joinResp := client1.GetSentMessages()[0].Data.(*JoinResponse)
		roomID := joinResp.RoomID

		// Create new client for reconnection
		client2 := client.NewClientMock("client5")
		reconnectMsg := CreateMessage("reconnect", map[string]interface{}{
			"clientId": client1.ID(),
			"roomID":   roomID,
		})
		router.HandleMessage(client2, reconnectMsg)

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

	t.Run("add bot flow", func(t *testing.T) {
		client1 := client.NewClientMock("client9")
		joinMsg := CreateMessage("join_room", map[string]interface{}{
			"gameType":   "testGame",
			"playerName": "tester-1",
		})
		router.HandleMessage(client1, joinMsg)

		// Clear messages from join
		client1.ClearMessages()

		addBotMsg := CreateMessage("add_bot", nil)
		router.HandleMessage(client1, addBotMsg)

		// Check the response
		messages := client1.GetSentMessages()
		if len(messages) < 1 {
			t.Errorf("expected at least 1 message after adding bot, got %d", len(messages))
			return
		}

		// Find the add_bot_result message
		var botResponse *protocol.Response
		for _, msg := range messages {
			if msg.Type == "add_bot_result" {
				botResponse = msg
				break
			}
		}

		if botResponse.Type == "" {
			t.Errorf("expected add_bot_result message not found in responses")
			return
		}

		if !botResponse.Success {
			t.Errorf("add_bot should succeed but got failure. Error: %s", botResponse.Error)
		}
	})

	t.Run("game action without room should fail", func(t *testing.T) {
		clientF := client.NewClientMock("client_no_room")

		msg := CreateMessage("game_action", map[string]interface{}{
			"action": "test_action",
		})
		router.HandleMessage(clientF, msg)

		messages := clientF.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 error message, got %d", len(messages))
		}

		response := messages[0]
		if response.Success != false {
			t.Errorf("expected game action to fail, got success")
		}
		if response.Type != "game_action_result" {
			t.Errorf("expected game_action_result, got %s", response.Type)
		}
	})

	t.Run("add bot without room should fail", func(t *testing.T) {
		clientF := client.NewClientMock("client_no_room_bot")

		msg := CreateMessage("add_bot", nil)
		router.HandleMessage(clientF, msg)

		messages := clientF.GetSentMessages()
		if len(messages) != 1 {
			t.Errorf("expected 1 error message, got %d", len(messages))
		}

		response := messages[0]
		if response.Success != false {
			t.Errorf("expected add bot to fail, got success")
		}
		if response.Type != "add_bot_result" {
			t.Errorf("expected add_bot_result, got %s", response.Type)
		}
	})
}
