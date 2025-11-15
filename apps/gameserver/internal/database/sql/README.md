# SQL Database Package

This package provides a generic SQL database interface for the gameserver, supporting both PostgreSQL and SQLite databases.

## Usage

## Connection Strings

### PostgreSQL

```
postgres://username:password@localhost:5432/database_name?sslmode=disable
postgresql://user:password@host:port/dbname?sslmode=require
```

### SQLite

```
file:./myapp.db
file:path/to/db.sqlite?cache=shared&mode=rwc
```

## Environment Variables

```bash
# PostgreSQL
DATABASE_URL=postgres://user:pass@localhost:5432/tellit
DATABASE_URL=postgresql://user:password@host:5432/db?sslmode=require

# SQLite
DATABASE_URL=file:./db.sqlite
DATABASE_URL=myapp.db
```

### Example

```go
// SQLite
db, _ := sql.New(ctx, "myapp.db")

// PostgreSQL
db, _ := sql.New(ctx, "postgres://user:pass@localhost/db?sslmode=disable")

// CRUD operations
db.Create(ctx, "users", user)
db.Get(ctx, "users", "id123", &user)
db.Update(ctx, "users", "id123", user)
db.Delete(ctx, "users", "id123")
db.Query(ctx, "SELECT * FROM users", &users)

// Transactions
tx, _ := db.BeginTx(ctx, nil)
tx.Create(ctx, "users", user)
tx.Commit() // or tx.Rollback()
```
