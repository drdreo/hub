# Game Server Integration Guide

## Data Flow

- Store client ID and room ID for reconnect in session storage, not local storage to avoid shared data issues between browser tabs.

### Client Events

Client can send these events to the server:

- `create_room`: Creates a new game room and joins it automatically
    - Payload: `{ gameType: string, options?: object }`
    - Response: `create_room_result`
- `join_room`: Joins an existing room by ID
    - Payload: `{ roomId: string }`
    - Response: `join_room_result`
- `leave_room`: Explicitly leave the current room
    - Payload: `{}`
    - Response: `leave_room_result`
- `reconnect`: Attempt to reconnect to a previously joined room
    - Payload: `{ roomId?: string, clientId: string }`
    - Response: `reconnect_result`
- `game_action`: Generic wrapper for game-specific actions (like `make_move`)
    - Payload: Varies by game
    - Response: `game_action_result`
    - not required to stick to `game_action`, any message will be routed to the game

### Server Events

Server sends these events to clients:

- `welcome`: Initial connection established
- `create_room_result`: Result when creating a room
    - Data: `{ roomId: string, gameType: string }`
- `join_room_result`: Result when joining a room
    - Data: `{ roomId: string, gameType: string, clients: number }`
- `leave_room_result`: Result when leaving a room
    - Data: `{ roomId: string }`
- `reconnect_result`: Result of reconnection attempt
    - Data: `{ gameType: string, roomId: string }`
- `client_joined`: Notification when another client joins the room
    - Data: `{ clientId: string }`
- `client_left`: Notification when another client leaves the room
    - Data: `{ clientId: string }`

### Game-Specific Events

Each game can implement these events to handle extra data synchronization:

- `joined`: Successfully joined a game room
    - Data: `{ clientId: string, roomId: string, symbol: string, ... }`
- `reconnected`: Successfully reconnected to a game room
    - Data: `{ clientId: string, roomId: string, symbol: string, ... }`
- `game_state`: Current state of the game (sent after every state change)
    - Data: `{ board: any, players: object, currentTurn: string, ... }`

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

2. **Testing**
    - Test reconnection scenarios extensively
    - Verify behavior with multiple clients

## Example: Minimal Client Setup

```javascript
class GameClient {
    constructor(serverUrl) {
        this.serverUrl = serverUrl;
        this.socket = null;
        this.clientId = sessionStorage.getItem('clientId');
        this.roomId = sessionStorage.getItem('roomId');
        this.eventHandlers = {};
    }

    connect() {
        this.socket = new WebSocket(this.serverUrl);

        this.socket.onopen = () => this.onConnectionOpen();
        this.socket.onclose = () => this.onConnectionClosed();
        this.socket.onerror = (error) => this.onConnectionError(error);
        this.socket.onmessage = (event) => this.onMessage(event);
    }

    onConnectionOpen() {
        console.log("Connected to game server");

        // Try to reconnect if we have previous session
        if (this.clientId && this.roomId) {
            this.reconnect(this.roomId, this.clientId);
        }

        this.trigger('connected');
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

- WebSocket connection handling using Gorilla WebSocket
- Client session tracking and lifecycle management
- Bidirectional communication with browser clients

### Room System

- Dynamic room creation and management
- Room joining/leaving logic
- Targeted message broadcasting (to specific clients or rooms)
- Room state persistence

### Message Routing

- Protocol-based message routing
- Game-specific message handling
- Efficient message distribution

### Game Registry

- Centralized registry for game implementations
- Dynamic loading of game logic
- Game-specific configuration and initialization

## Architecture Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                      Game Server (Go)                          │
│                                                                │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────────────┐   │
│  │  WebSocket  │   │    Room     │   │   Game Registry     │   │
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
// Client interface
type Client interface {
ID() string
Send(message []byte) error
Room() Room
SetRoom(room Room)
Close()
}

// Room interface
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
}

// Game interface
type Game interface {
Type() string
HandleMessage(client Client, room Room, msgType string, payload []byte)
InitializeRoom(room Room, options json.RawMessage) error
OnClientJoin(client Client, room Room)
OnClientLeave(client Client, room Room)
}
```

## Message Flow

1. Client connects to server via WebSocket
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
    - Client reconnects (browser loaded)

2. Client sends "reconnect" message with old client ID from localStorage

    - Server fetches session data from global store
    - Server rejoins client to the room
    - Game handles reconnection logic via OnClientReconnect
    - Session is removed from the store

3. Cleanup routine automatically removes sessions after timeout
