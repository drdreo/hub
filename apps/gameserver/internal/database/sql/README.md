# SQL Database Package

This package provides a generic SQL database interface for the gameserver, supporting both PostgreSQL and SQLite databases.

## Usage

### Connection Strings

**PostgreSQL:**

```
postgres://user:password@host:port/dbname?sslmode=disable
```

**SQLite:**

```
file:path/to/db.sqlite?cache=shared&mode=rwc
```

or simply:

```
path/to/db.sqlite
```

### Example

```go
// SQLite
client, _ := sql.NewClient(ctx, "myapp.db")

// PostgreSQL
client, _ := sql.NewClient(ctx, "postgres://user:pass@localhost/db?sslmode=disable")

// CRUD operations
client.Create(ctx, "users", user)
client.Get(ctx, "users", "id123", &user)
client.Update(ctx, "users", "id123", user)
client.Delete(ctx, "users", "id123")
client.Query(ctx, "SELECT * FROM users", &users)

// Transactions
tx, _ := client.BeginTx(ctx, nil)
tx.Create(ctx, "users", user)
tx.Commit() // or tx.Rollback()
```


## Environment Variables

Use `DATABASE_URL` to provide the connection string:

```bash
# PostgreSQL
DATABASE_URL=postgres://user:pass@localhost:5432/tellit

# SQLite
DATABASE_URL=file:./db.sqlite
```
