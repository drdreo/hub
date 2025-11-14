package sql

import (
	"context"
	"database/sql"
)

type Database interface {
	Create(ctx context.Context, table string, data interface{}) error
	Get(ctx context.Context, table string, id string, dest interface{}) error
	Update(ctx context.Context, table string, id string, data interface{}) error
	Delete(ctx context.Context, table string, id string) error
	Query(ctx context.Context, query string, dest interface{}, args ...interface{}) error
	// Exec executes a query without returning any rows
	Exec(ctx context.Context, query string, args ...interface{}) error
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
}

type Tx interface {
	Create(ctx context.Context, table string, data interface{}) error
	Update(ctx context.Context, table string, id string, data interface{}) error
	Delete(ctx context.Context, table string, id string) error
	Query(ctx context.Context, query string, dest interface{}, args ...interface{}) error
	// Exec executes a query without returning any rows within a transaction
	Exec(ctx context.Context, query string, args ...interface{}) error
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
}
