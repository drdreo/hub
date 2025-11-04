# Game Server Architecture Diagram

## Current Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Game Server (Go)                             │
│                                                                     │
│  ┌──────────────────┐    ┌──────────────────┐    ┌──────────────┐ │
│  │ Client Manager   │    │  Room Manager    │    │ Game Registry│ │
│  │                  │    │                  │    │              │ │
│  │ - Register       │◄───┤ - Create Room    │◄───┤ - TicTacToe  │ │
│  │ - Unregister     │    │ - Get Room       │    │ - Dice Game  │ │
│  │ - Get by Game    │    │ - Remove Room    │    │ - Owe Drahn  │ │
│  └────────┬─────────┘    └────────┬─────────┘    └──────┬───────┘ │
│           │                       │                      │         │
│  ┌────────▼─────────┐    ┌────────▼─────────┐    ┌──────▼───────┐ │
│  │  WebSocket       │    │  Game Room       │    │ Game         │ │
│  │  Client          │◄───┤  - ID            │◄───┤ Interface    │ │
│  │  - ID            │    │  - Clients       │    │ - Handle Msg │ │
│  │  - Send          │    │  - State         │    │ - Init Room  │ │
│  │  - Read Pump     │    │  - Join/Leave    │    │ - On Join    │ │
│  │  - Write Pump    │    │  - Broadcast     │    │ - On Leave   │ │
│  └──────────────────┘    └──────────────────┘    └──────────────┘ │
│           ▲                       ▲                                │
│           │                       │                                │
│  ┌────────┴─────────┐    ┌────────┴─────────┐                     │
│  │  Message Router  │    │  Session Store   │                     │
│  │  - Route Msgs    │    │  - Store Session │                     │
│  │  - Handle Join   │    │  - Get Session   │                     │
│  │  - Handle Leave  │    │  - Cleanup       │                     │
│  │  - Reconnect     │    │  (In-Memory)     │                     │
│  └──────────────────┘    └──────────────────┘                     │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                            ▲           ▲           ▲
                            │           │           │
                    ┌───────┴───┐  ┌────┴────┐  ┌──┴──────┐
                    │ Browser   │  │ Browser │  │ Browser │
                    │ Client 1  │  │ Client 2│  │ Client 3│
                    └───────────┘  └─────────┘  └─────────┘
```

## Message Flow

### 1. Client Connection

```
Browser                 Server
  │                       │
  ├──WebSocket Connect───►│
  │                       ├─Create WebSocketClient
  │                       ├─Register in ClientManager
  │◄────Welcome Msg───────┤
  │                       │
```

### 2. Join Room

```
Browser                 Server
  │                       │
  ├──join_room───────────►│
  │                       ├─Get/Create Room
  │                       ├─Game.InitializeRoom
  │                       ├─Room.Join(client)
  │                       ├─Game.OnClientJoin
  │◄──join_room_result───┤
  │◄──game_state─────────┤
  │                       │
```

### 3. Game Action

```
Browser                 Server
  │                       │
  ├──make_move───────────►│
  │                       ├─Game.HandleMessage
  │                       ├─Validate Move
  │                       ├─Update State
  │                       ├─Room.Broadcast
  │◄──game_state─────────┤ (to all players)
  │                       │
```

### 4. Reconnection

```
Browser                 Server
  │                       │
  ├──reconnect───────────►│
  │ {clientId, roomId}    ├─SessionStore.GetSession
  │                       ├─Verify Session Valid
  │                       ├─Room.Join (new client)
  │                       ├─Game.OnClientReconnect
  │◄──reconnect_result───┤
  │◄──game_state─────────┤ (current state)
  │                       │
```

## Data Structures

### Room State

```go
Room {
    id:       string
    gameType: string
    clients:  map[string]Client
    state:    interface{}  // ⚠️ Mutable reference (Issue #1)
    closed:   bool
    mu:       sync.RWMutex
}
```

### Session Data

```go
SessionData {
    ClientID:  string
    RoomID:    string
    GameType:  string
    LeftAt:    time.Time
    ExtraData: map[string]interface{}
}

// ⚠️ Stored in-memory only (Issue #2)
SessionStore {
    sessions: map[string]SessionData
}
```

### Message Protocol

```go
// Client → Server
Message {
    Type:     string           // "join_room", "make_move", etc.
    RoomID:   string           (optional)
    GameType: string           (optional)
    Data:     json.RawMessage  // Payload
}

// Server → Client
Response {
    Type:    string      // "game_state", "error", etc.
    Success: bool
    Error:   string      // ⚠️ Should be structured (Issue #4)
    Data:    interface{}
}
```

## Concurrency Model

### WebSocket Client Goroutines

```
┌─────────────────────────────────────┐
│          WebSocket Client           │
│                                     │
│  ┌────────────┐    ┌─────────────┐ │
│  │ Read Pump  │    │ Write Pump  │ │
│  │ goroutine  │    │ goroutine   │ │
│  │            │    │             │ │
│  │ Reads msgs │    │ Sends msgs  │ │
│  │ from conn  │    │ to conn     │ │
│  │            │    │             │ │
│  │    │       │    │   ▲         │ │
│  └────┼───────┘    └───┼─────────┘ │
│       │                │           │
│       ▼                │           │
│  OnMessage          send chan      │
│   Handler            (buffered)    │
└─────────────────────────────────────┘
```

### Room Broadcast

```
Room.Broadcast(msg)
    │
    ├─For each client in room
    │   │
    │   └─client.Send(msg)
    │       │
    │       └─send chan ← msg  // ⚠️ Can drop if full (Issue #3)
    │
    └─All sends complete
```

## Scalability Concerns

### Current: Single Instance

```
                 ┌─────────────┐
                 │   Clients   │
                 └──────┬──────┘
                        │
                 ┌──────▼──────┐
                 │  Game       │
                 │  Server     │
                 │  Instance   │
                 └──────┬──────┘
                        │
                 ┌──────▼──────┐
                 │  In-Memory  │
                 │  Sessions   │
                 └─────────────┘
```

✅ Works well  
❌ Cannot scale horizontally  
❌ Single point of failure

### Recommended: Multi-Instance with Redis

```
                 ┌─────────────┐
                 │   Clients   │
                 └──────┬──────┘
                        │
                 ┌──────▼──────┐
            ┌────┤     Load    ├────┐
            │    │   Balancer  │    │
            │    └─────────────┘    │
            │                       │
     ┌──────▼──────┐         ┌──────▼──────┐
     │  Game       │         │  Game       │
     │  Server 1   │         │  Server 2   │
     └──────┬──────┘         └──────┬──────┘
            │                       │
            └───────────┬───────────┘
                        │
                 ┌──────▼──────┐
                 │    Redis    │
                 │  Cluster    │
                 │  (Sessions) │
                 └─────────────┘
```

✅ Horizontal scaling  
✅ High availability  
✅ Session persistence

## Thread Safety Issues

### Issue #1: State Race Condition

```go
// ❌ CURRENT (Unsafe)
func (room *GameRoom) State() interface{} {
    room.mu.RLock()
    defer room.mu.RUnlock()
    return room.state  // Returns mutable reference!
}

// Game code can do:
state := room.State().(GameState)
state.Board[0][0] = "X"  // ⚠️ Race condition! No lock held!

// ✅ RECOMMENDED (Safe)
// Make state immutable
func (room *GameRoom) State() interface{} {
    room.mu.RLock()
    defer room.mu.RUnlock()
    return room.state  // OK if games never mutate
}

// Games must do:
oldState := room.State().(GameState)
newState := oldState.WithMove(0, 0, "X")  // Creates new state
room.SetState(newState)  // Atomic update with lock
```

## Performance Characteristics

### Message Batching (Implemented ✅)

```
Without Batching:
┌────┐ ┌────┐ ┌────┐
│Msg1│ │Msg2│ │Msg3│  3 WebSocket frames
└────┘ └────┘ └────┘

With Batching:
┌──────────────────┐
│[Msg1, Msg2, Msg3]│  1 WebSocket frame
└──────────────────┘

Result: 3x reduction in frame overhead
```

### Room Cleanup Lock Contention (Issue #5)

```
Every 5 minutes:

❌ CURRENT:
Lock all rooms (write lock)
Check all rooms
Delete empty rooms
Unlock

⏸️ All operations blocked during cleanup

✅ RECOMMENDED:
Lock rooms (read lock)
Collect empty room IDs
Unlock
For each empty room:
    Lock that specific room
    Delete if still empty
    Unlock

✅ Operations only blocked for individual deletions
```

## Security Model

### CORS Protection

```
Production:
  Origin: *.drdreo.com ✅

Development:
  Origin: * ✅
```

### Message Size Limits

```
Max WebSocket Message: 2048 bytes ✅
Max Send Buffer: 256 messages ✅
```

### Missing Protections

```
❌ No rate limiting per client
❌ No authentication/authorization
❌ No input sanitization framework
❌ No connection limits per IP
```

## Testing Strategy

### Current Test Coverage

```
✅ Unit Tests:
   - Game logic (TicTacToe, Dice)
   - Session store
   - Room management

✅ Integration Tests:
   - Full game flow
   - Router message handling

❌ Missing:
   - Load tests (1000+ concurrent)
   - Concurrency tests (go test -race)
   - Chaos tests (network failures)
   - Performance benchmarks
```

## Summary

**Architecture Quality: 7.5/10**

✅ **Strengths:**

-   Clean interface design
-   Plugin-based extensibility
-   Good WebSocket implementation
-   Working concurrency primitives

⚠️ **Needs Improvement:**

-   Fix race conditions (#1)
-   Add horizontal scaling (#2)
-   Handle message drops (#3)
-   Structure error responses (#4)
-   Optimize cleanup (#5)

**Recommendation:**
Fix critical issues (#1, #2) before production scale deployment.
The architecture is sound but needs specific improvements for reliability and scalability.
