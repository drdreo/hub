package room

import (
	"gameserver/internal/client"
	"gameserver/internal/protocol"
	"testing"
)

func TestRoom(t *testing.T) {
	t.Run("client_join_behavior", func(t *testing.T) {
		// Create mock clients
		client1 := client.NewClientMock("client1")
		client2 := client.NewClientMock("client2")

		// Create room and add clients
		room := NewRoom("testGame", nil)
		room.Join(client1)

		// At this point, client1 shouldn't have messages since it's the first to join
		if len(client1.GetSentMessages()) != 0 {
			t.Errorf("first client received message on empty room join, got %d messages",
				len(client1.GetSentMessages()))
		}

		room.Join(client2)

		// Now client1 should receive notification about client2 joining
		client1Messages := client1.GetSentMessages()
		if len(client1Messages) != 1 {
			t.Errorf("client1 got %d messages after client2 joined, expected 1",
				len(client1Messages))
		}

		// client2 shouldn't receive its own join message
		client2Messages := client2.GetSentMessages()
		if len(client2Messages) != 0 {
			t.Errorf("client2 got %d messages about its own join, expected 0",
				len(client2Messages))
		}

		// Verify the join message content
		if len(client1Messages) > 0 {
			msg := client1Messages[0]

			// Check message type
			if msg.Type != "client_joined" {
				t.Errorf("expected message type 'client_joined', got '%v'", msg.Type)
			}

			// Check client ID
			if data, ok := msg.Data.(map[string]interface{}); !ok || data["clientId"] != "client2" {
				t.Errorf("expected clientId 'client2', got '%v'", data["clientId"])
			}
		}
	})

	t.Run("broadcast_behavior", func(t *testing.T) {
		// Create mock clients
		client1 := client.NewClientMock("client1")
		client2 := client.NewClientMock("client2")
		client3 := client.NewClientMock("client3")

		// Create room and add clients
		room := NewRoom("testGame", nil)
		room.Join(client1)
		room.Join(client2)
		room.Join(client3)

		client1.ClearMessages()
		client2.ClearMessages()
		client3.ClearMessages()

		testMessage := protocol.NewSuccessResponse("test", "hello")
		room.Broadcast(testMessage, client1) // client1 is the sender

		// Verify the message wasn't sent to the sender
		if len(client1.GetSentMessages()) != 0 {
			t.Errorf("sender client received message, expected no messages")
		}

		client2Messages := client2.GetSentMessages()
		client3Messages := client3.GetSentMessages()

		// Check message count for client2
		if len(client2Messages) != 1 {
			t.Errorf("client2 got %d messages, expected 1", len(client2Messages))
		}

		// Check message count for client3
		if len(client3Messages) != 1 {
			t.Errorf("client3 got %d messages, expected 1", len(client3Messages))
		}

		// Check message content for client2
		if len(client2Messages) > 0 && client2Messages[0] != testMessage {
			t.Errorf("client2 got message %s, expected %s", client2Messages[0].Data, testMessage.Data)
		}

		// Check message content for client3
		if len(client3Messages) > 0 && client3Messages[0] != testMessage {
			t.Errorf("client3 got message %s, expected %s", client3Messages[0].Data, testMessage.Data)
		}
	})

	t.Run("client_leave_behavior", func(t *testing.T) {
		client1 := client.NewClientMock("client1")
		client2 := client.NewClientMock("client2")

		room := NewRoom("testGame", nil)
		room.Join(client1)
		room.Join(client2)

		client1.ClearMessages()
		client2.ClearMessages()

		room.Leave(client2)

		// Now client1 should receive notification about client2 leaving
		client1Messages := client1.GetSentMessages()
		if len(client1Messages) != 1 {
			t.Errorf("client1 got %d messages after client2 left, expected 1",
				len(client1Messages))
		}

		// Verify the leave message content
		if len(client1Messages) > 0 {
			msg := client1Messages[0]

			// Check message type
			if msg.Type != "client_left" {
				t.Errorf("expected message type 'client_left', got '%v'", msg.Type)
			}

			// Check client ID
			if data, ok := msg.Data.(map[string]interface{}); !ok || data["clientId"] != "client2" {
				t.Errorf("expected clientId 'client2', got '%v'", data["clientId"])
			}
		}

		// Check if client2 is removed from the room
		if _, exists := room.clients[client2.ID()]; exists {
			t.Errorf("client2 still exists in the room after leaving")
		}
	})

	t.Run("room_close_behavior", func(t *testing.T) {
		client1 := client.NewClientMock("client1")
		client2 := client.NewClientMock("client2")

		room := NewRoom("testGame", nil)
		room.Join(client1)
		room.Join(client2)

		// Clear messages before closing the room
		client1.ClearMessages()
		client2.ClearMessages()

		// Close the room
		room.Close()

		// Verify that the room is closed
		if !room.IsClosed() {
			t.Errorf("room is not closed after calling Close()")
		}

		// Now both clients should receive notification about the room closing
		client1Messages := client1.GetSentMessages()
		client2Messages := client2.GetSentMessages()

		if len(client1Messages) != 1 {
			t.Errorf("client1 got %d messages after room close, expected 1",
				len(client1Messages))
		}

		if len(client2Messages) != 1 {
			t.Errorf("client2 got %d messages after room close, expected 1",
				len(client2Messages))
		}

		// Verify the close message content
		if len(client1Messages) > 0 {
			msg := client1Messages[0]

			// Check message type
			if msg.Type != "room_closed" {
				t.Errorf("expected message type 'room_closed', got '%v'", msg.Type)
			}

			// Check room ID
			if data, ok := msg.Data.(map[string]interface{}); !ok || data["roomId"] != room.ID() {
				t.Errorf("expected roomId '%s', got '%v'", room.ID(), data["roomId"])
			}
		}

		if len(client2Messages) > 0 {
			msg := client2Messages[0]

			// Check message type
			if msg.Type != "room_closed" {
				t.Errorf("expected message type 'room_closed', got '%v'", msg.Type)
			}

			// Check room ID
			if data, ok := msg.Data.(map[string]interface{}); !ok || data["roomId"] != room.ID() {
				t.Errorf("expected roomId '%s', got '%v'", room.ID(), data["roomId"])
			}
		}
	})
}
