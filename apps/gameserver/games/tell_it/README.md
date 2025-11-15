# Tell-It Game

A collaborative storytelling game where players take turns adding text to create stories together.

## Overview

Tell-It is a real-time multiplayer game where players:

1. Join a room together
2. Start the game when ready (minimum 2 players)
3. Take turns adding text to stories
4. Stories circulate between players, with each player adding to different stories
5. Vote to finish the game when stories are complete
6. View all completed stories at the end

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

### Configuration

Set the `TELLIT_DATABASE_URL` environment variable:

```bash
# SQLite (development)
TELLIT_DATABASE_URL=file:./db.sqlite

# PostgreSQL (production)
TELLIT_DATABASE_URL=postgres://user:password@host:port/dbname
```

If not set, defaults to SQLite in development mode.

### Schema

```sql
CREATE TABLE stories (
    id TEXT PRIMARY KEY,
    room_name TEXT NOT NULL,
    text TEXT NOT NULL,
    author TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_stories_room ON stories(room_name);
```
