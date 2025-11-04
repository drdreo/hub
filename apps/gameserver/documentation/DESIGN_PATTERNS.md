# Game Server Architecture - Design Patterns & Best Practices

**Generated:** November 4, 2025  
**Related:** [ARCHITECTURE_REVIEW.md](./ARCHITECTURE_REVIEW.md) | [ACTION_ITEMS.md](./ARCHITECTURE_ACTION_ITEMS.md)

---

## Design Patterns in Use

### ✅ Current Good Patterns

#### 1. **Strategy Pattern** (Game Interface)

```go
// Allows pluggable game implementations
type Game interface {
    Type() string
    HandleMessage(client Client, room Room, msgType string, data []byte) error
    InitializeRoom(ctx context.Context, room Room, options json.RawMessage) error
    OnClientJoin(client Client, room Room, options CreateRoomOptions)
    OnClientLeave(client Client, room Room)
}

// Usage
tictactoe.RegisterTicTacToeGame(gameRegistry)
dicegame.RegisterDiceGame(gameRegistry)
```

**Benefits:**

-   Easy to add new games
-   No modification to core server code
-   Clean separation of concerns

---

#### 2. **Interface Segregation Principle**

```go
// Small, focused interfaces
type Client interface {
    ID() string
    Send(message *protocol.Response) error
    Room() Room
    SetRoom(room Room)
    Close()
}

type Room interface {
    ID() string
    Join(client Client) error
    Broadcast(message *protocol.Response, exclude ...Client)
    // ... other room-specific methods
}
```

**Benefits:**

-   Testability (easy to mock)
-   Flexibility (multiple implementations)
-   Clear contracts

---

#### 3. **Dependency Injection**

```go
// Dependencies passed through constructor
func NewRouter(
    ctx context.Context,
    clientManager interfaces.ClientManager,
    roomManager interfaces.RoomManager,
    gameRegistry interfaces.GameRegistry,
) *Router {
    return &Router{
        ctx:           ctx,
        clientManager: clientManager,
        roomManager:   roomManager,
        gameRegistry:  gameRegistry,
    }
}
```

**Benefits:**

-   Testable without globals
-   Configurable dependencies
-   Clear dependency graph

---

#### 4. **Command Pattern** (Message Routing)

```go
// Messages as commands
switch message.Type {
case "join_room":
    r.handleJoinRoom(ctx, client, message.Data)
case "leave_room":
    r.handleLeaveRoom(client)
case "reconnect":
    r.handleReconnect(client, message.Data)
// ...
}
```

**Benefits:**

-   Extensible message types
-   Centralized routing logic
-   Easy to add new commands

---

### 🔄 Recommended Pattern Improvements

#### 5. **Immutable State Pattern** (Recommended)

```go
// CURRENT (mutable, unsafe):
func (room *GameRoom) State() interface{} {
    room.mu.RLock()
    defer room.mu.RUnlock()
    return room.state  // Mutable reference!
}

// RECOMMENDED (immutable, safe):
type GameState struct {
    // Make all fields private
    board [3][3]string
    // ...
}

func (s GameState) WithMove(row, col int, symbol string) GameState {
    // Return new state with change
    newState := s
    newState.board[row][col] = symbol
    return newState
}

// In game logic:
oldState := room.State().(GameState)
newState := oldState.WithMove(row, col, "X")
room.SetState(newState)
```

**Benefits:**

-   Thread-safe by design
-   No race conditions
-   Easier to reason about
-   Time-travel debugging possible

---

#### 6. **Repository Pattern** (Recommended for Games)

```go
// Current: Each game implements own storage
// Recommended: Shared repository interface

type GameRepository interface {
    SaveGame(ctx context.Context, roomID string, state interface{}) error
    LoadGame(ctx context.Context, roomID string) (interface{}, error)
    DeleteGame(ctx context.Context, roomID string) error
    ListGames(ctx context.Context, gameType string) ([]string, error)
}

// Implementations:
// - MemoryRepository (for testing)
// - FirestoreRepository (for production)
// - SQLRepository (alternative)

// Usage in game:
type PersistentGame struct {
    repo GameRepository
}

func (g *PersistentGame) OnClientLeave(client Client, room Room) {
    // Auto-save game state
    g.repo.SaveGame(context.Background(), room.ID(), room.State())
}
```

---

#### 7. **Observer Pattern** (For Game Events)

```go
// Recommended: Allow external systems to observe game events

type GameEventType string

const (
    GameStarted GameEventType = "game_started"
    GameEnded   GameEventType = "game_ended"
    PlayerJoined GameEventType = "player_joined"
)

type GameEvent struct {
    Type     GameEventType
    RoomID   string
    GameType string
    Data     interface{}
}

type GameObserver interface {
    OnGameEvent(event GameEvent)
}

// Use cases:
// - Analytics tracking
// - Achievement systems
// - Leaderboard updates
// - Push notifications
```

---

#### 8. **Builder Pattern** (For Complex Creation)

```go
// Current: Many parameters in CreateRoom
type CreateRoomOptions struct {
    GameType   string
    PlayerName string
    RoomID     *string
    Options    json.RawMessage
}

// Recommended: Builder for better API
type RoomBuilder struct {
    gameType   string
    playerName string
    roomID     *string
    options    map[string]interface{}
    maxPlayers int
    private    bool
}

func NewRoomBuilder(gameType string) *RoomBuilder {
    return &RoomBuilder{
        gameType:   gameType,
        maxPlayers: 2,
        private:    false,
    }
}

func (b *RoomBuilder) WithPlayer(name string) *RoomBuilder {
    b.playerName = name
    return b
}

func (b *RoomBuilder) WithCustomID(id string) *RoomBuilder {
    b.roomID = &id
    return b
}

func (b *RoomBuilder) MaxPlayers(n int) *RoomBuilder {
    b.maxPlayers = n
    return b
}

func (b *RoomBuilder) Private() *RoomBuilder {
    b.private = true
    return b
}

func (b *RoomBuilder) Build(ctx context.Context) (*Room, error) {
    // Validation and construction
}

// Usage:
room, err := NewRoomBuilder("tictactoe").
    WithPlayer("Alice").
    MaxPlayers(2).
    Build(ctx)
```

---

## Concurrency Patterns

### ✅ Currently Implemented

#### Read-Write Mutex for State

```go
type GameRoom struct {
    mu sync.RWMutex
    // ...
}

func (room *GameRoom) Clients() map[string]Client {
    room.mu.RLock()  // Multiple readers OK
    defer room.mu.RUnlock()
    return maps.Clone(room.clients)
}

func (room *GameRoom) Join(client Client) error {
    room.mu.Lock()  // Exclusive write access
    defer room.mu.Unlock()
    // ...
}
```

**Good!** Allows concurrent reads while protecting writes.

---

#### Worker Goroutines

```go
// WebSocket read/write pumps
func (c *WebSocketClient) StartPumps() {
    go c.writePump()  // Dedicated writer
    go c.readPump()   // Dedicated reader
}
```

**Good!** Separates concerns and prevents blocking.

---

### 🔄 Recommended Additions

#### Fan-Out/Fan-In for Broadcasting

```go
// Current: Sequential broadcast
func (room *GameRoom) Broadcast(message *protocol.Response, exclude ...Client) {
    for _, client := range room.clients {
        client.Send(message)  // Sequential
    }
}

// Recommended: Parallel broadcast for many clients
func (room *GameRoom) BroadcastParallel(message *protocol.Response, exclude ...Client) {
    excludeMap := make(map[string]bool)
    for _, c := range exclude {
        excludeMap[c.ID()] = true
    }

    var wg sync.WaitGroup
    for _, client := range room.clients {
        if !excludeMap[client.ID()] {
            wg.Add(1)
            go func(c Client) {
                defer wg.Done()
                c.Send(message)
            }(client)
        }
    }
    wg.Wait()  // Wait for all sends to complete
}
```

**Use when:** Broadcasting to 10+ clients

---

#### Context for Cancellation

```go
// Recommended: Use context for graceful shutdown

type Router struct {
    ctx    context.Context
    cancel context.CancelFunc
    // ...
}

func (r *Router) Shutdown() {
    r.cancel()  // Signal all goroutines to stop
}

// In background workers:
func (s *Store) cleanupRoutine() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            s.cleanup()
        case <-r.ctx.Done():  // Graceful shutdown
            log.Info().Msg("Cleanup routine stopping")
            return
        }
    }
}
```

---

## Error Handling Patterns

### ✅ Current Approach

```go
// Errors sent to client
client.Send(protocol.NewErrorResponse("error", "Room not found"))
```

### 🔄 Recommended: Structured Errors

```go
// Define error types
type ErrorCode string

const (
    ErrRoomNotFound     ErrorCode = "ROOM_NOT_FOUND"
    ErrGameFull         ErrorCode = "GAME_FULL"
    ErrInvalidMove      ErrorCode = "INVALID_MOVE"
    ErrNotYourTurn      ErrorCode = "NOT_YOUR_TURN"
    ErrSessionExpired   ErrorCode = "SESSION_EXPIRED"
)

type GameError struct {
    Code    ErrorCode
    Message string
    Context map[string]interface{}
}

func (e GameError) Error() string {
    return string(e.Code) + ": " + e.Message
}

// Usage:
if room == nil {
    return &GameError{
        Code:    ErrRoomNotFound,
        Message: "The requested room does not exist",
        Context: map[string]interface{}{
            "roomId": roomID,
        },
    }
}

// Convert to protocol response
func ErrorToResponse(err error) *protocol.Response {
    if gameErr, ok := err.(*GameError); ok {
        return protocol.NewErrorResponse("error", &protocol.ErrorDetail{
            Code:    string(gameErr.Code),
            Message: gameErr.Message,
            Details: gameErr.Context,
        })
    }
    return protocol.NewErrorResponse("error", err.Error())
}
```

---

## Testing Patterns

### Recommended Test Structure

```go
// Table-driven tests
func TestCalculateScore(t *testing.T) {
    tests := []struct {
        name     string
        dice     []int
        expected int
        valid    bool
    }{
        {"empty_dice", []int{}, 0, false},
        {"single_1", []int{1}, 100, true},
        {"triple_ones", []int{1, 1, 1}, 1000, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            score, valid := calculateScore(tt.dice)
            if score != tt.expected || valid != tt.valid {
                t.Errorf("got %d, %v; want %d, %v",
                    score, valid, tt.expected, tt.valid)
            }
        })
    }
}
```

### Mock Interfaces

```go
// Mock client for testing
type MockClient struct {
    id       string
    room     interfaces.Room
    messages []*protocol.Response
}

func (m *MockClient) Send(msg *protocol.Response) error {
    m.messages = append(m.messages, msg)
    return nil
}

// Usage in tests
func TestRoomBroadcast(t *testing.T) {
    client1 := &MockClient{id: "c1"}
    client2 := &MockClient{id: "c2"}

    room := NewRoom(nil, "test", nil)
    room.Join(client1)
    room.Join(client2)

    room.Broadcast(protocol.NewSuccessResponse("test", nil))

    if len(client1.messages) != 1 {
        t.Error("Expected 1 message")
    }
}
```

---

## Performance Patterns

### Message Batching (Already Implemented!)

```go
// apps/gameserver/internal/client/socket_client.go:201-211
w.Write([]byte("["))
w.Write(message)

// Add queued messages to the current websocket message
n := len(c.send)
for i := 0; i < n; i++ {
    w.Write([]byte(","))
    w.Write(<-c.send)
}
w.Write([]byte("]"))
```

**Excellent pattern!** Reduces WebSocket frames and improves throughput.

---

### Object Pooling (Recommended for High Load)

```go
// For frequently allocated objects
var messagePool = sync.Pool{
    New: func() interface{} {
        return &protocol.Response{}
    },
}

func GetResponse() *protocol.Response {
    return messagePool.Get().(*protocol.Response)
}

func PutResponse(r *protocol.Response) {
    // Reset fields
    r.Type = ""
    r.Success = false
    r.Error = ""
    r.Data = nil
    messagePool.Put(r)
}

// Usage:
resp := GetResponse()
resp.Type = "game_state"
resp.Success = true
resp.Data = state
client.Send(resp)
PutResponse(resp)
```

---

## Security Patterns

### Input Validation

```go
// Always validate and sanitize
type JoinRoomRequest struct {
    PlayerName string `json:"playerName" validate:"required,min=1,max=50,alphanum"`
    RoomID     string `json:"roomId" validate:"omitempty,uuid"`
}

// Validate before processing
if err := validate.Struct(request); err != nil {
    return ErrInvalidInput
}
```

### Rate Limiting (Recommended)

```go
// Per-client rate limiter
type RateLimiter struct {
    mu      sync.Mutex
    clients map[string]*rate.Limiter
}

func (rl *RateLimiter) Allow(clientID string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.clients[clientID]
    if !exists {
        // 10 messages per second per client
        limiter = rate.NewLimiter(10, 20)
        rl.clients[clientID] = limiter
    }

    return limiter.Allow()
}

// Usage in message handler:
if !rateLimiter.Allow(client.ID()) {
    client.Send(protocol.NewErrorResponse("error", "Rate limit exceeded"))
    return
}
```

---

## Documentation Patterns

### Code Documentation

```go
// Good: Clear purpose and usage
// HandleMessage routes incoming WebSocket messages to appropriate handlers.
// It validates the message format, checks client authorization, and forwards
// to game-specific handlers or built-in commands.
//
// Thread-safety: Safe for concurrent calls from multiple goroutines.
// Each client's messages are processed sequentially by their read pump.
func (r *Router) HandleMessage(client interfaces.Client, messageData []byte) {
    // ...
}
```

### Architecture Decision Records (ADRs)

Create `docs/adr/` folder with decisions:

```
docs/adr/
  001-websocket-over-http-polling.md
  002-plugin-based-game-architecture.md
  003-go-for-backend-implementation.md
```

---

## Recommended Go Packages

### Already Using ✅

-   `github.com/gorilla/websocket` - WebSocket support
-   `github.com/rs/zerolog` - Structured logging
-   `github.com/google/uuid` - UUID generation

### Recommended Additions 🔄

-   `github.com/go-playground/validator/v10` - Input validation
-   `github.com/prometheus/client_golang` - Metrics
-   `github.com/go-redis/redis/v8` - Redis client
-   `golang.org/x/time/rate` - Rate limiting
-   `go.opentelemetry.io/otel` - Distributed tracing

---

## Summary

**Excellent patterns already in use:**

1. Interface-driven design
2. Dependency injection
3. Strategy pattern for games
4. Message batching optimization
5. Proper concurrency primitives

**Recommended additions:**

1. Immutable state pattern
2. Repository pattern for storage
3. Observer pattern for events
4. Structured error handling
5. Context-based cancellation
6. Rate limiting

**Next Steps:**

1. Review the HIGH PRIORITY items in [ACTION_ITEMS.md](./ARCHITECTURE_ACTION_ITEMS.md)
2. Implement immutable state pattern first
3. Add observability (metrics + tracing)
4. Prepare for horizontal scaling

---

**Remember:** These are recommendations, not requirements. Prioritize based on your specific needs and timeline.
