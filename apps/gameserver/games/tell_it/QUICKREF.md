# Tell-It Game - Quick Reference

## Game Type

```
"tellit"
```

## WebSocket Connection

```
ws://localhost:8080/ws                    # Development
wss://hub.drdreo.com/ws                   # Production
```

## Room Configuration

```json
{
    "gameType": "tellit",
    "playerName": "Alice",
    "options": {
        "spectatorsAllowed": false,
        "isPublic": true,
        "minUsers": 2,
        "maxUsers": 8,
        "afkDelay": 30000
    }
}
```

## Client Messages

### Start Game

```json
{
    "type": "start"
}
```

### Submit Text

```json
{
    "type": "submit_text",
    "data": {
        "text": "Once upon a time..."
    }
}
```

### Vote to Finish

```json
{
    "type": "vote_finish"
}
```

### Vote to Restart

```json
{
    "type": "vote_restart"
}
```

### Vote to Kick

```json
{
    "type": "vote_kick",
    "data": {
        "kickUserID": "user-id-here"
    }
}
```

### Request Stories

```json
{
    "type": "request_stories"
}
```

### Request Update

```json
{
    "type": "request_update"
}
```

## Server Messages

### Joined

```json
{
    "type": "joined",
    "success": true,
    "data": {
        "userID": "abc123",
        "room": "room-name"
    }
}
```

### Users Update

```json
{
    "type": "users_update",
    "success": true,
    "data": {
        "users": [
            {
                "id": "user1",
                "name": "Alice",
                "disconnected": false,
                "afk": false,
                "kickVotes": [],
                "queuedStories": 1
            }
        ]
    }
}
```

### Game Status

```json
{
    "type": "game_status",
    "success": true,
    "data": {
        "status": "started" // "waiting" | "started" | "ended"
    }
}
```

### Story Update

```json
{
    "type": "story_update",
    "success": true,
    "data": {
        "text": "Once upon a time",
        "author": "Bob"
    }
}
```

### Finish Vote Update

```json
{
    "type": "finish_vote_update",
    "success": true,
    "data": {
        "votes": ["Alice", "Bob"]
    }
}
```

### Final Stories

```json
{
    "type": "final_stories",
    "success": true,
    "data": {
        "stories": [
            {
                "text": "Once upon a time. There was a knight. He fought bravely.",
                "author": "Alice"
            }
        ]
    }
}
```

### User Left

```json
{
    "type": "user_left",
    "success": true,
    "data": {
        "userID": "user123"
    }
}
```

### User Kicked

```json
{
    "type": "user_kicked",
    "success": true,
    "data": {
        "kickedUser": "BadPlayer"
    }
}
```

### Error

```json
{
    "type": "submit_text",
    "success": false,
    "error": "can't wait - no story to continue"
}
```

## Game Flow

1. **Room Creation**: Create room with game type "tellit"
2. **Players Join**: 2-8 players join the room
3. **Start Game**: Any player can start once min players reached
4. **Submit Text**: Players take turns adding to stories
5. **Story Circulation**: Stories rotate between players
6. **Vote Finish**: All players vote when stories complete
7. **View Stories**: All completed stories displayed
8. **Vote Restart** (optional): All players vote to play again

## Story Circulation Pattern

```
User A creates Story 1 → submits text → Story 1 goes to User B
User B receives Story 1 → submits text → Story 1 goes to User C
User C receives Story 1 → submits text → Story 1 goes to User A
User A receives Story 1 → continues...
```
