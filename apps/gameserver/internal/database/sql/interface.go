package sql

import (
	"context"
	"database/sql"
)

// Repository defines CRUD operations for database records
type Repository interface {
	Create(ctx context.Context, table string, data interface{}) error
	Get(ctx context.Context, table string, id string, dest interface{}) error
	Update(ctx context.Context, table string, id string, data interface{}) error
	Delete(ctx context.Context, table string, id string) error
}

// Querier defines custom query execution capabilities
type Querier interface {
	Query(ctx context.Context, query string, dest interface{}, args ...interface{}) error
	// Exec executes a query without returning any rows
	Exec(ctx context.Context, query string, args ...interface{}) error
}

// TransactionManager defines transaction lifecycle management
type TransactionManager interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Transaction, error)
}

// TransactionControl defines transaction control operations
type TransactionControl interface {
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
}

// Transaction defines all operations that can be performed within a database transaction
type Transaction interface {
	Repository
	Querier
	TransactionControl
}

// Database defines all database operations including transaction management
type Database interface {
	Repository
	Querier
	TransactionManager
	// Close closes the database connection
	Close() error
	// Driver returns the name of the database driver (e.g., "postgres", "sqlite")
	Driver() string
}

