package room

import (
	"bytes"
	"encoding/json"
	"gameserver/internal/client"
	"testing"
)

func TestClientJoin(t *testing.T) {
	// Create mock clients
	client1 := client.NewClientMock("client1")
	client2 := client.NewClientMock("client2")

	// Create room and add clients
	room := NewRoom("testGame")
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
		var msg map[string]interface{}
		if err := json.Unmarshal(client1Messages[0], &msg); err != nil {
			t.Errorf("failed to parse join message: %v", err)
		} else {
			// Check message type
			if msgType, ok := msg["type"].(string); !ok || msgType != "client_joined" {
				t.Errorf("expected message type 'client_joined', got '%v'", msg["type"])
			}

			// Check client ID
			if clientID, ok := msg["clientId"].(string); !ok || clientID != "client2" {
				t.Errorf("expected clientId 'client2', got '%v'", msg["clientId"])
			}
		}
	}
}

func TestRoomBroadcast(t *testing.T) {
	// Create mock clients
	client1 := client.NewClientMock("client1")
	client2 := client.NewClientMock("client2")
	client3 := client.NewClientMock("client3")

	// Create room and add clients
	room := NewRoom("testGame")
	room.Join(client1)
	room.Join(client2)
	room.Join(client3)

	client1.ClearMessages()
	client2.ClearMessages()
	client3.ClearMessages()

	// Test broadcasting
	testMessage := []byte(`{"type":"test","data":"hello"}`)
	room.Broadcast(testMessage, client1) // client1 is the sender

	// Verify the message wasnt sent to the sender
	if len(client1.GetSentMessages()) != 0 {
		t.Errorf("sender client received message, expected no messages")
	}

	client2Messages := client2.GetSentMessages()
	client3Messages := client3.GetSentMessages()

	// Check message count for client2
	if len(client2Messages) != 1 {
		t.Errorf("client2 got %d messages, expected 1", len(client2Messages))
	}

	// Check message count for client2
	if len(client3Messages) != 1 {
		t.Errorf("client3 got %d messages, expected 1", len(client3Messages))
	}

	// Check message content for client2
	if len(client2Messages) > 0 && !bytes.Equal(client2Messages[0], testMessage) {
		t.Errorf("client2 got message %s, expected %s", client2Messages[0], testMessage)
	}

	// Check message content for client3
	if len(client3Messages) > 0 && !bytes.Equal(client3Messages[0], testMessage) {
		t.Errorf("client3 got message %s, expected %s", client3Messages[0], testMessage)
	}
}
