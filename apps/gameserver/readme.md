# Game Server Integration Guide

## 📚 Documentation

**Start Here:** [Architecture Review Summary](../../ARCHITECTURE_REVIEW_SUMMARY.md) (Repository Root)

### Architecture Documentation

-   **[Architecture Review](./documentation/ARCHITECTURE_REVIEW.md)** (21KB) - Comprehensive architectural analysis with detailed findings
-   **[Action Items](./documentation/ARCHITECTURE_ACTION_ITEMS.md)** (8KB) - Prioritized list of improvements with effort estimates
-   **[Design Patterns](./documentation/DESIGN_PATTERNS.md)** (15KB) - Best practices and design patterns with code examples
-   **[Architecture Diagram](./documentation/ARCHITECTURE_DIAGRAM.md)** (13KB) - Visual diagrams, flow charts, and data structures

## Data Flow

-   Store client ID and room ID for reconnect in session storage, not local storage to avoid shared data issues between
    browser tabs.

### Client Events

Client can send these events to the server (standard request messages). All messages use the envelope `{ type: string, data?: any }`.

-   `join_room`
    -   Purpose: Join an existing room or create one if `roomId` is omitted or not found.
    -   Payload (`data`): `{ gameType: string, roomId?: string, playerName: string, options?: any }`
    -   Success Response: `join_room_result` with data `{ clientId: string, roomId: string }`
    -   Error Response: `join_room_result` with `success: false` and `error` message
    -   Special Case: If already in a room and you send `join_room`, the server returns a `reconnect_result` success instead (auto treat as reconnect) containing `{ clientId, roomId }`.
-   `leave_room`
    -   Purpose: Leave the current room.
    -   Payload: `{}` (no data needed)
    -   Success Response: `leave_room_result` with `data: null`
    -   Error Response: `leave_room_result` with `error` if not in a room
-   `reconnect`
    -   Purpose: Re-associate a new socket with a previous session.
    -   Payload: `{ clientId: string, roomId?: string }`
    -   Success Response: `reconnect_result` with data `{ clientId: string, roomId: string, gameType: string }`
    -   Error Response: `reconnect_result` with `error` (invalid session, room not found, etc.)
-   `game_action`
    -   Purpose: Generic wrapper to send a game-specific action payload (the game sees `type = game_action`).
    -   Payload: Game-defined JSON
    -   Success/Error Response: `game_action_result`
    -   Note: You do NOT have to use `game_action`; any other `type` sent while in a room is forwarded directly to the game's handler.
    -   Direct Game Actions (Forwarded without wrapper)
        - Examples (TicTacToe): `make_move`, `restart_game`
        - Response: Game typically emits updated `game_state` or specific events; errors come back as `error` messages.
-   `add_bot`
    -   Purpose: Add a bot player to the current room (if supported by the game).
    -   Payload: `{}`
    -   Success Response: `add_bot_result` (data is `null`)
    -   Error Response: `add_bot_result` with `error`
-   `get_room_list`
    -   Purpose: Request current list of rooms for a given game type.
    -   Payload: `{ gameType: string }`
    -   Success Response: `room_list_update` (see below) immediately for requester
    -   Error Response: `get_room_list_result` with `error`
-   Other / Unknown Types
    -   If the client is in a room, unknown types are passed to the game's `HandleMessage`; if not in a room, you get `error`.

### Server Events

Server sends these events to clients:

-   `welcome`
    -   Sent on initial WebSocket connection (before joining a room)
    -   Data: `{ message: string }`
-   `join_room_result`
    -   Data (success): `{ clientId: string, roomId: string }`
    -   Data (error): `error` string
-   `leave_room_result`
    -   Data (success): `null`
    -   Data (error): `error` string
-   `reconnect_result`
    -   Data (success): `{ clientId: string, roomId: string, gameType: string }`
    -   Data (error): `error` string
    -   Note: When using `join_room` while already in a room, a success `reconnect_result` (without `gameType`) is returned to facilitate seamless UX.
-   `client_joined`
    -   Data: `{ clientId: string }` (broadcast to other clients when someone joins)
-   `client_left`
    -   Data: `{ clientId: string }` (broadcast when someone leaves)
-   `room_closed`
    -   Data: `{ roomId: string }` (broadcast when room is closed)
-   `room_list_update`
    -   Data: `Array<{ roomId: string, playerCount: number, started: boolean }>` (pushed on changes and on `get_room_list` success)
-   `add_bot_result`
    -   Data (success): `null`
    -   Data (error): `error` string
-   `get_room_list_result`
    -   Only sent on error for the `get_room_list` client action with `error` message (success uses `room_list_update`).
-   `error`
    -   Data: `{ success: false, error: string }` (generic validation & unknown message errors)

### Game-Specific Events (Optional / Per-Game)

Each game may define additional events; not all games emit the same ones. For example:

-   TicTacToe:
    -   `joined`: `{ clientId, symbol, roomId }` (sent only to the joining client)
    -   `reconnected`: `{ clientId, symbol, roomId }` (sent only to the reconnecting client)
    -   `game_state`: Full board & player state `{ board, players, currentTurn, winner, gameOver, drawGame }` (after each change)
-   DiceGame:
    -   `game_state`: `{ players, dice, selectedDice, setAside, started, currentTurn, winner, targetScore, ... }`
    -   `busted`: `{ clientId, name }` (after a player busts a roll; may be delayed for animation)

### Response Format

All server responses share a common shape:

```
{ type: string, success: boolean, error?: string, data?: any }
```

Use `success` to drive optimistic UI updates; if `success` is false check `error`.

## Implementation Tips

### Client Side

1. **Connection Management**

    - Store connection data in session storage with a TTL
    - Implement automatic reconnection with exponential backoff
    - Handle connection errors gracefully with user feedback

2. **State Handling**

    - Keep local state in sync with server state
        - Especially after reconnecting
    - Events can be grouped into one message. Parse the message and handle all received events.

3. **Reconnection Logic**
    - Always check game state after reconnection due to possible race conditions
    - If reconnected and game_state events arrive out of order, use the most recent

### Server Side

1. **Game Implementation**

    - Ensure all state changes are atomic
    - Use proper locking for concurrent access to shared resources
    - Register game handlers through the game registry

## Example: Minimal Client Setup

```javascript
class GameClient {
    constructor(serverUrl) {
        this.serverUrl = serverUrl;
        this.socket = null;
        this.clientId = sessionStorage.getItem("clientId");
        this.roomId = sessionStorage.getItem("roomId");
        this.eventHandlers = {};
    }

    connect() {
        this.socket = new WebSocket(this.serverUrl);

        this.socket.onopen = () => this.onConnectionOpen();
        this.socket.onclose = () => this.onConnectionClosed();
        this.socket.onerror = error => this.onConnectionError(error);
        this.socket.onmessage = event => this.onMessage(event);
    }

    onConnectionOpen() {
        console.log("Connected to game server");

        // Try to reconnect if we have previous session
        if (this.clientId && this.roomId) {
            this.reconnect(this.roomId, this.clientId);
        }

        this.trigger("connected");
    }

    // Additional methods...

    // Event handling system
    on(event, callback) {
        if (!this.eventHandlers[event]) {
            this.eventHandlers[event] = [];
        }
        this.eventHandlers[event].push(callback);
    }

    trigger(event, data) {
        const handlers = this.eventHandlers[event] || [];
        handlers.forEach(handler => handler(data));
    }
}
```

## Common Issues & Solutions

1. **Stale UI**
    - Problem: UI doesn't reflect current game state after reconnect
    - Solution: Force UI refresh on reconnection completion

# Game Server Architecture

This document outlines the architecture of the WebSocket-based game server implemented in Go, designed to support
multiple room-based multiplayer browser games.

## Overview

The game server provides a centralized infrastructure for managing WebSocket connections, rooms, and game sessions. It
uses a plugin-based architecture that allows different games to register their logic with the server while sharing
common infrastructure.

## Core Components

### Connection Management

-   WebSocket connection handling using Gorilla WebSocket
-   Client session tracking and lifecycle management

### Room System

-   Room creation and management
-   Room joining/leaving/reconnecting logic
-   Targeted message broadcasting (to specific clients or rooms)
-   Room state persistence

### Message Routing

-   Protocol-based message routing
-   Game-specific message handling
-   Efficient message distribution

### Game Registry

-   Centralized registry for game implementations
-   Game-specific configuration and initialization

## Architecture Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                      Game Server (Go)                          │
│                                                                │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────────────┐   │
│  │  Client     │   │    Room     │   │   Game Registry     │   │
│  │  Manager    │◄──┤   Manager   │◄──┤                     │   │
│  └─────┬───────┘   └─────┬───────┘   │  ┌───────────────┐  │   │
│        │                 │           │  │ Chess Game    │  │   │
│  ┌─────▼───────┐   ┌─────▼───────┐   │  └───────────────┘  │   │
│  │   Client    │   │    Room     │   │  ┌───────────────┐  │   │
│  │   Sessions  │◄──┤   Sessions  │◄──┤  │ Poker Game    │  │   │
│  └─────────────┘   └─────────────┘   │  └───────────────┘  │   │
│                                      │  ┌───────────────┐  │   │
│                                      │  │ Trivia Game   │  │   │
│                                      │  └───────────────┘  │   │
│                                      └─────────────────────┘   │
│                                                                │
└────────────────────────────────────────────────────────────────┘
             │                  │                  │
             │                  │                  │
             ▼                  ▼                  ▼
   ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
   │  Browser Client │ │  Browser Client │ │  Browser Client │
   └─────────────────┘ └─────────────────┘ └─────────────────┘
```

## Key Interfaces

```go
package docs

import "encoding/json"

// interfaces/interfaces.go

type Client interface {
	ID() string
	Send(message []byte) error
	Room() Room
	SetRoom(room Room)
	Close()
}

type Room interface {
	ID() string
	GameType() string
	Join(client Client) error
	Leave(client Client)
	Broadcast(message []byte, exclude ...Client)
	BroadcastTo(message []byte, clients ...Client)
	Clients() map[string]Client
	State() interface{}
	SetState(state interface{})
	Close()
	IsClosed() bool
}

type Game interface {
	Type() string
	HandleMessage(client Client, room Room, msgType string, payload []byte)
	InitializeRoom(room Room, options json.RawMessage) error
	OnClientJoin(client Client, room Room)
	OnClientLeave(client Client, room Room)
}

type RoomManager interface {
	CreateRoom(options interface{}) (Room, error)
	GetRoom(roomID string) (Room, error)
	RemoveRoom(roomID string)
	ListRooms() []Room
	GetAllRoomsByGameType(gameType string) []Room
}

type ClientManager interface {
	RegisterClient(client Client, gameType string)
	UnregisterClient(client Client)
	GetClientsByGameType(gameType string) []Client
}

```

## Message Flow

1. Client connects to server via WebSocket (should provide game type)
2. Client joins or creates a room with specific game type
3. Server initializes room with game-specific logic
4. Client sends game actions as messages
5. Server routes messages to appropriate game handler
6. Game logic processes messages and updates room state
7. Server broadcasts state changes to clients in room

## Reconnection Flow

1. Client disconnects (browser refresh)

    - client.Close() stores session data in global session store
    - Session includes client ID, room ID, game type, and metadata

2. Client reconnects (on browser load)

    - Client sends "reconnect" message with old client ID from sessionStorage
    - Server fetches session data from global store
    - Server rejoins client to the room
    - Game handles reconnection logic via OnClientReconnect
    - Session is removed from the store

3. Cleanup routine automatically removes sessions after timeout
