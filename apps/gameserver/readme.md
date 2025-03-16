# Game Server Architecture

This document outlines the architecture of the WebSocket-based game server implemented in Go, designed to support multiple room-based multiplayer browser games.

## Overview

The game server provides a centralized infrastructure for managing WebSocket connections, rooms, and game sessions. It uses a plugin-based architecture that allows different games to register their logic with the server while sharing common infrastructure.

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

## Client Integration

The server provides a framework-agnostic client API that can be integrated with any frontend:

```typescript
// Example client usage
const connection = new GameConnection("ws://localhost:8080/ws");

connection
  .on("room_joined", (data) => {
    console.log(`Joined room: ${data.roomId}`);
  })
  .on("game_state", (state) => {
    // Update UI with new game state
    updateGameUI(state);
  });

// Join a room
connection.joinRoom("room123", { gameType: "chess" });

// Send game action
connection.send("move", { from: "e2", to: "e4" });
```

## Implementation Plan

1. Core WebSocket server setup
2. Client and room management implementation
3. Message routing and protocol definition
4. Game registry and plugin system
5. Sample game implementation
6. Client library development
7. Testing and optimization

## Benefits

- Centralized infrastructure for multiple games
- Shared connection handling and room management
- Efficient resource usage
- Easy addition of new games
- Framework-agnostic client API